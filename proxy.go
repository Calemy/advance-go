package main

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
)

var proxy *ProxyRotator

type ProxyRotator struct {
	proxies []*url.URL
	counter uint64
}

func NewProxyRotator(proxyStrings []string) (*ProxyRotator, error) {
	var parsed []*url.URL

	for _, p := range proxyStrings {
		u, err := url.Parse(p)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, u)
	}

	return &ProxyRotator{proxies: parsed}, nil
}

func (r *ProxyRotator) NextProxy(_ *http.Request) (*url.URL, error) {
	i := atomic.AddUint64(&r.counter, 1)

	p := r.proxies[i%uint64(len(r.proxies))]
	return p, nil
}

func initProxies() {
	proxies := make([]string, 0)

	if os.Getenv("ENABLE_PROXY") == "true" {
		file, err := os.OpenFile("proxy.txt", os.O_CREATE|os.O_RDONLY, 0600)
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			proxies = append(proxies, scanner.Text())
		}

		log.Printf("Successfully registered %d proxies", len(proxies))
	}

	rotator, err := NewProxyRotator(proxies)
	if err != nil {
		panic(err)
	}
	proxy = rotator
}
