package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type ScoresResponse struct {
	Scores       []Score `json:"scores"`
	CursorString *string `json:"cursor_string"`
}

type Score struct { //We are currently using lazer metrics and yet not all of the ones listed below
	Accuracy          float64         `json:"accuracy"`
	BeatmapID         int             `json:"beatmap_id"`
	BuildID           *int     package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type ScoresResponse struct {
	Scores       []Score `json:"scores"`
	CursorString *string `json:"cursor_string"`
}

type Score struct { //We are currently using lazer metrics and yet not all of the ones listed below
	Accuracy          float64         `json:"accuracy"`
	BeatmapID         int             `json:"beatmap_id"`
	BuildID           *int            `json:"build_id"`
	ClassicTotalScore int             `json:"classic_total_score"`
	EndedAt           time.Time       `json:"ended_at"`
	HasReplay         bool            `json:"has_replay"`
	ID                int             `json:"id"`
	IsPerfectCombo    bool            `json:"is_perfect_combo"`
	LegacyPerfect     bool            `json:"legacy_perfect"`
	LegacyScoreID     *int            `json:"legacy_score_id"`
	LegacyTotalScore  int             `json:"legacy_total_score"`
	MaxCombo          int             `json:"max_combo"`
	MaximumStatistics ScoreStatistics `json:"maximum_statistics"`
	Mods              []Mod           `json:"mods"`
	Passed            bool            `json:"passed"`
	PlaylistItemID    *int            `json:"playlist_item_id"`
	PP                float64         `json:"pp"`
	Preserve          *bool           `json:"preserve"`
	Processed         *bool           `json:"processed"`
	Rank              string          `json:"rank"`
	Ranked            *bool           `json:"ranked"`
	RoomID            *int            `json:"room_id"`
	RulesetID         int             `json:"ruleset_id"`
	StartedAt         *time.Time      `json:"started_at"`
	Statistics        ScoreStatistics `json:"statistics"`
	TotalScore        int             `json:"total_score"`
	Type              string          `json:"type"`
	UserID            int             `json:"user_id"`
}

type ScoreStatistics struct { //Most of them are currently unused
	Miss                int `json:"miss"`
	Meh                 int `json:"meh"`
	Ok                  int `json:"ok"`
	Good                int `json:"good"`
	Great               int `json:"great"`
	Perfect             int `json:"perfect"`
	SmallTickMiss       int `json:"small_tick_miss"`
	SmallTickHit        int `json:"small_tick_hit"`
	LargeTickMiss       int `json:"large_tick_miss"`
	LargeTickHit        int `json:"large_tick_hit"`
	SmallBonus          int `json:"small_bonus"`
	LargeBonus          int `json:"large_bonus"`
	IgnoreMiss          int `json:"ignore_miss"`
	IgnoreHit           int `json:"ignore_hit"`
	ComboBreak          int `json:"combo_break"`
	SliderTailHit       int `json:"slider_tail_hit"`
	LegacyComboIncrease int `json:"legacy_combo_increase"`
}

func fetchScores() {
	data, err := Fetch("/scores?cursor_string=" + cursor)
	if err != nil {
		if err == ErrFetch {
			log.Println("Error while fetching scores endpoint.")
			return
		}
	}

	if len(data) == 0 {
		log.Println("Empty response body, skipping")
		return
	}

	var scores ScoresResponse

	if err := json.Unmarshal(data, &scores); err != nil {
		log.Println("Something went wrong while unmarshling")
		os.WriteFile("error.txt", []byte(err.Error()), 0644)
		os.WriteFile("data.json", data, 0644)
		return //Once this fires it just freezes?
	}

	start := time.Now()

	defer func() {
		log.Printf("%d scores inserted in %s", len(scores.Scores), time.Since(start))
	}()

	if len(scores.Scores) == 0 {
		return
	}

	var wg sync.WaitGroup

	wg.Add(len(scores.Scores))

	for _, score := range scores.Scores {
		go func(s Score) {
			defer wg.Done()
			if !userCache.Exists(s.UserID) { //move to create?
				user := &UserExtended{ID: s.UserID}
				if err := user.Create(); err == nil {
					userCache.Add(s.UserID)
				}
			}

			playedCache.Add(s.UserID)
			s.Insert()
		}(score)
	}
	wg.Wait()

	if scores.CursorString != nil { //Moved it down here so it only writes cursor after inserts to not skip any by accient
		cursor = *scores.CursorString
		if err := os.WriteFile("cursor.txt", []byte(cursor), 0644); err != nil {
			log.Println("Couldn't write cursor to file?", err.Error())
		}
	}
}

func (s *Score) Insert() {
	if _, err := DB.Exec(context.Background(), `
		INSERT INTO scores_go (
			user_id,
			beatmap,
			score_id,
			score,
			accuracy,
			max_combo,
			count_50,
			count_100,
			count_300,
			count_miss,
			fc,
			mods,
			time,
			rank,
			passed,
			pp,
			mode,
			added
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, 
			$15, $16, $17, $18
		) ON CONFLICT (score_id) DO NOTHING
		`,
		s.UserID,
		s.BeatmapID,
		s.ID,
		s.TotalScore,
		s.Accuracy*100,
		s.MaxCombo,
		s.Statistics.Meh,
		s.Statistics.Ok,
		s.Statistics.Great,
		s.Statistics.Miss,
		s.IsPerfectCombo,
		convertMods(s.Mods),
		s.EndedAt,
		s.Rank,
		s.Passed,
		s.PP,
		s.RulesetID,
		time.Now(),
	); err != nil {
		log.Println("Something went wrong inserting score", err)
		return
	}

	scoreCount++
}
       `json:"build_id"`
	ClassicTotalScore int             `json:"classic_total_score"`
	EndedAt           time.Time       `json:"ended_at"`
	HasReplay         bool            `json:"has_replay"`
	ID                int             `json:"id"`
	IsPerfectCombo    bool            `json:"is_perfect_combo"`
	LegacyPerfect     bool            `json:"legacy_perfect"`
	LegacyScoreID     *int            `json:"legacy_score_id"`
	LegacyTotalScore  int             `json:"legacy_total_score"`
	MaxCombo          int             `json:"max_combo"`
	MaximumStatistics ScoreStatistics `json:"maximum_statistics"`
	Mods              []Mod           `json:"mods"`
	Passed            bool            `json:"passed"`
	PlaylistItemID    *int            `json:"playlist_item_id"`
	PP                float64         `json:"pp"`
	Preserve          *bool           `json:"preserve"`
	Processed         *bool           `json:"processed"`
	Rank              string          `json:"rank"`
	Ranked            *bool           `json:"ranked"`
	RoomID            *int            `json:"room_id"`
	RulesetID         int             `json:"ruleset_id"`
	StartedAt         *time.Time      `json:"started_at"`
	Statistics        ScoreStatistics `json:"statistics"`
	TotalScore        int             `json:"total_score"`
	Type              string          `json:"type"`
	UserID            int             `json:"user_id"`
}

type ScoreStatistics struct { //Most of them are currently unused
	Miss                int `json:"miss"`
	Meh                 int `json:"meh"`
	Ok                  int `json:"ok"`
	Good                int `json:"good"`
	Great               int `json:"great"`
	Perfect             int `json:"perfect"`
	SmallTickMiss       int `json:"small_tick_miss"`
	SmallTickHit        int `json:"small_tick_hit"`
	LargeTickMiss       int `json:"large_tick_miss"`
	LargeTickHit        int `json:"large_tick_hit"`
	SmallBonus          int `json:"small_bonus"`
	LargeBonus          int `json:"large_bonus"`
	IgnoreMiss          int `json:"ignore_miss"`
	IgnoreHit           int `json:"ignore_hit"`
	ComboBreak          int `json:"combo_break"`
	SliderTailHit       int `json:"slider_tail_hit"`
	LegacyComboIncrease int `json:"legacy_combo_increase"`
}

func fetchScores() {
	data, err := Fetch("/scores?cursor_string=" + cursor)
	if err != nil {
		if err == ErrFetch {
			log.Println("Error while fetching scores endpoint.")
			return
		}
	}

	var scores ScoresResponse

	if err := json.Unmarshal(data, &scores); err != nil {
		log.Println("Something went wrong while unmarshling")
		os.WriteFile("error.txt", []byte(err.Error()), 0644)
		return
	}

	start := time.Now()

	defer func() {
		log.Printf("%d scores inserted in %s", len(scores.Scores), time.Since(start))
	}()

	if len(scores.Scores) == 0 {
		return
	}

	var wg sync.WaitGroup

	wg.Add(len(scores.Scores))

	for _, score := range scores.Scores {
		go func(s Score) {
			defer wg.Done()
			if !userCache.Exists(s.UserID) { //move to create?
				user := &UserExtended{ID: s.UserID}
				if err := user.Create(); err == nil {
					userCache.Add(s.UserID)
				}
			}

			playedCache.Add(s.UserID)
			s.Insert()
		}(score)
	}
	wg.Wait()

	if scores.CursorString != nil { //Moved it down here so it only writes cursor after inserts to not skip any by accient
		cursor = *scores.CursorString
		if err := os.WriteFile("cursor.txt", []byte(cursor), 0644); err != nil {
			log.Println("Couldn't write cursor to file?", err.Error())
		}
	}
}

func (s *Score) Insert() {
	if _, err := DB.Exec(context.Background(), `
		INSERT INTO scores_go (
			user_id,
			beatmap,
			score_id,
			score,
			accuracy,
			max_combo,
			count_50,
			count_100,
			count_300,
			count_miss,
			fc,
			mods,
			time,
			rank,
			passed,
			pp,
			mode,
			added
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, 
			$15, $16, $17, $18
		) ON CONFLICT (score_id) DO NOTHING
		`,
		s.UserID,
		s.BeatmapID,
		s.ID,
		s.TotalScore,
		s.Accuracy*100,
		s.MaxCombo,
		s.Statistics.Meh,
		s.Statistics.Ok,
		s.Statistics.Great,
		s.Statistics.Miss,
		s.IsPerfectCombo,
		convertMods(s.Mods),
		s.EndedAt,
		s.Rank,
		s.Passed,
		s.PP,
		s.RulesetID,
		time.Now(),
	); err != nil {
		log.Println("Something went wrong inserting score", err)
		return
	}

	scoreCount++
}
