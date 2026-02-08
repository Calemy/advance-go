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

	Beatmap    *BeatmapExtended `json:"beatmap"`
	Beatmapset *Beatmapset      `json:"beatmapset"`
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

type BeatmapExtended struct {
	ID           int `json:"id"`
	BeatmapsetID int `json:"beatmapset_id"`
	UserID       int `json:"user_id"`

	ModeInt int `json:"mode_int"`

	Ranked int `json:"ranked"`

	Version          string  `json:"version"`
	DifficultyRating float64 `json:"difficulty_rating"`

	TotalLength int `json:"total_length"`
	HitLength   int `json:"hit_length"`

	Accuracy float64 `json:"accuracy"`
	AR       float64 `json:"ar"`
	CS       float64 `json:"cs"`
	Drain    float64 `json:"drain"`

	BPM *float64 `json:"bpm"`

	Playcount int `json:"playcount"`
	Passcount int `json:"passcount"`

	LastUpdated time.Time `json:"last_updated"`
}

func (b *BeatmapExtended) Insert(s *Beatmapset) error {
	_, err := DB.Exec(context.Background(), `
		INSERT INTO beatmaps (
			beatmap_id,
			beatmapset_id,
			title,
			artist,
			creator,
			creator_id,
			version,
			length,
			ranked,
			last_update,
			added
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			NOW(),
			NOW()
		)
		ON CONFLICT (beatmap_id) DO UPDATE
		SET
			beatmapset_id = EXCLUDED.beatmapset_id,
			title         = EXCLUDED.title,
			artist        = EXCLUDED.artist,
			creator       = EXCLUDED.creator,
			creator_id    = EXCLUDED.creator_id,
			version       = EXCLUDED.version,
			length        = EXCLUDED.length,
			ranked        = EXCLUDED.ranked,
			last_update = CASE
				WHEN (
					beatmaps.beatmapset_id IS DISTINCT FROM EXCLUDED.beatmapset_id OR
					beatmaps.title         IS DISTINCT FROM EXCLUDED.title OR
					beatmaps.artist        IS DISTINCT FROM EXCLUDED.artist OR
					beatmaps.creator       IS DISTINCT FROM EXCLUDED.creator OR
					beatmaps.creator_id    IS DISTINCT FROM EXCLUDED.creator_id OR
					beatmaps.version       IS DISTINCT FROM EXCLUDED.version OR
					beatmaps.length        IS DISTINCT FROM EXCLUDED.length OR
					beatmaps.ranked        IS DISTINCT FROM EXCLUDED.ranked
				)
				THEN NOW()
				ELSE beatmaps.last_update
			END;
	`,
		b.ID,
		s.ID,
		s.Title,
		s.Artist,
		s.Creator,
		b.UserID,
		b.Version,
		b.HitLength,
		b.Ranked,
	)

	return err
}

type Beatmapset struct {
	ID int `json:"id"`

	Artist  string `json:"artist"`
	Title   string `json:"title"`
	Creator string `json:"creator"`

	Status string `json:"status"`

	UserID int `json:"user_id"`
}

const (
	ModeStd   uint8 = iota // 0
	ModeTaiko              // 1
	ModeCatch              // 2
	ModeMania              // 3
)

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
	newUsers := 0
	lastTime := time.Now()

	if len(scores.Scores) == 0 {
		return
	}

	defer func() {
		log.Printf("%d scores inserted in %s | %d new users queued (%d total) | remaining ratelimit: %d", len(scores.Scores), time.Since(start), newUsers, len(userCache.m), client.remoteRL.remaining)
		log.Printf("Queue: %d | Priority: %d | Total: %d", len(userUpdater.in), len(userUpdater.priority), len(userUpdater.cache))
		log.Printf("last scoretime: %s", lastTime.String())
	}()

	var wg sync.WaitGroup

	wg.Add(len(scores.Scores))

	for _, score := range scores.Scores {
		go func(s Score) {
			defer wg.Done()
			priority := false
			if !userCache.Exists(s.UserID) { //move to create?
				user := &UserExtended{ID: s.UserID}
				if err := user.Create(); err == nil {
					newUsers++
					userCount++
					userCache.Add(s.UserID)
					priority = true
				}
			}

			go userUpdater.Queue(s.UserID, uint8(s.RulesetID), priority)
			s.Insert()
		}(score)
		lastTime = score.EndedAt
	}
	wg.Wait()

	if scores.CursorString != nil { //Moved it down here so it only writes cursor after inserts to not skip any by accient
		cursor = *scores.CursorString
		if err := os.WriteFile("cursor.txt", []byte(cursor), 0644); err != nil {
			log.Println("Couldn't write cursor to file?", err.Error())
		}
	}
}

func (s *Score) Insert() error {
	if _, exists := scoreCache.Get(s.ID); exists {
		return nil
	}
	if _, err := DB.Exec(context.Background(), `
		INSERT INTO scores (
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
		return err
	}

	scoreCount++
	return nil
}
