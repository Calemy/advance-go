package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var token *string = nil
var tokenMut sync.Mutex

type authToken struct {
	Token string `json:"token"`
}

var client = NewLimitedClient(5)

func login() {
	tokenMut.Lock()
	defer tokenMut.Unlock()
	payload := map[string]string{
		"username": os.Getenv("BANCHO_USERNAME"),
		"password": os.Getenv("BANCHO_PASSWORD"),
	}

	details, _ := json.Marshal(payload)

	resp, err := http.Post("https://auth.catboy.best/login", "application/json", bytes.NewBuffer(details))

	if err != nil {
		panic("Authentication not reachable.") //Please for the love of god implement a fallback
		//TODO: Implement natively
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Authentication not reachable. Status: %d", resp.StatusCode)) //Please for the love of god implement a fallback
	}

	body, _ := io.ReadAll(resp.Body)

	var result = &authToken{}
	_ = json.Unmarshal(body, result)

	token = &result.Token
}

func Request(url string) (*http.Response, error) {
	tokenMut.Lock()
	if token == nil {
		tokenMut.Unlock()
		login()
	} else {
		tokenMut.Unlock()
	}

	var lastErr error

	for i := 0; i < 3; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))
		req.Header.Set("User-Agent", "osu-lazer")
		req.Header.Set("scope", "*")

		resp, err := client.Do(req)

		if err != nil {
			if strings.Contains(err.Error(), "server sent GOAWAY") {
				lastErr = err
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return nil, err
		}

		if resp.StatusCode == http.StatusUnauthorized {
			login()
			return Request(url)
		}

		return resp, nil
	}

	return nil, lastErr
}

func Fetch(endpoint string) ([]byte, error) {
	resp, err := Request(fmt.Sprintf("https://osu.ppy.sh/api/v2%s", endpoint))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, ErrNotFound
		}

		return nil, ErrFetch
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}
