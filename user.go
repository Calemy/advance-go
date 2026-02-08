package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type UserExtended struct {
	ID            int        `json:"id"`
	Username      string     `json:"username"`
	JoinDate      *time.Time `json:"join_date"`
	CountryCode   string     `json:"country_code"`
	AvatarURL     string     `json:"avatar_url"`
	IsActive      bool       `json:"is_active"`
	IsOnline      bool       `json:"is_online"`
	IsSupporter   bool       `json:"is_supporter"`
	LastVisit     *time.Time `json:"last_visit"`
	PMFriendsOnly bool       `json:"pm_friends_only"`
	ProfileColor  *string    `json:"profile_colour"`

	// Extended fields
	AccountHistory           []AccountHistory  `json:"account_history"`
	ActiveTournamentBanner   *TournamentBanner `json:"active_tournament_banner"`
	Badges                   []Badge           `json:"badges"`
	BeatmapPlaycountsCount   int               `json:"beatmap_playcounts_count"`
	FavouriteBeatmapsetCount int               `json:"favourite_beatmapset_count"`
	FollowerCount            int               `json:"follower_count"`
	GraveyardBeatmapsetCount int               `json:"graveyard_beatmapset_count"`
	Groups                   []UserGroup       `json:"groups"`
	LovedBeatmapsetCount     int               `json:"loved_beatmapset_count"`
	MappingFollowerCount     int               `json:"mapping_follower_count"`
	MonthlyPlaycounts        []MonthlyCount    `json:"monthly_playcounts"`
	Page                     UserPage          `json:"page"`
	PendingBeatmapsetCount   int               `json:"pending_beatmapset_count"`
	PreviousUsernames        []string          `json:"previous_usernames"`
	RankHighest              *UserRankHighest  `json:"rank_highest"`
	RankHistory              *RankHistory      `json:"rank_history"`
	RankedBeatmapsetCount    int               `json:"ranked_beatmapset_count"`
	ReplaysWatchedCounts     []MonthlyCount    `json:"replays_watched_counts"`

	ScoresBestCount   int `json:"scores_best_count"`
	ScoresFirstCount  int `json:"scores_first_count"`
	ScoresRecentCount int `json:"scores_recent_count"`

	Statistics         *UserStatistics            `json:"statistics"`
	StatisticsRulesets *map[string]UserStatistics `json:"statistics_rulesets"`
	SupportLevel       int                        `json:"support_level"`
	UserAchievements   []UserAchievement          `json:"user_achievements"`
}

type AccountHistory struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Length    int       `json:"length"`
}

type TournamentBanner struct {
	ID       int    `json:"id"`
	ImageURL string `json:"image_url"`
}

type Badge struct {
	AwardedAt   time.Time `json:"awarded_at"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	URL         string    `json:"url"`
}

type UserGroup struct {
	ID          int     `json:"id"`
	Identifier  string  `json:"identifier"`
	IsProbation bool    `json:"is_probation"`
	Name        string  `json:"name"`
	ShortName   string  `json:"short_name"`
	Colour      *string `json:"colour"`
}

type MonthlyCount struct {
	StartDate DateOnly `json:"start_date"`
	Count     int      `json:"count"`
}

type UserPage struct {
	HTML string `json:"html"`
	Raw  string `json:"raw"`
}

type UserRankHighest struct {
	Rank      int       `json:"rank"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RankHistory struct {
	Mode string `json:"mode"`
	Data []int  `json:"data"`
}

type UserStatistics struct {
	Level                  UserLevel     `json:"level"`
	GlobalRank             *int          `json:"global_rank"`
	CountryRank            *int          `json:"country_rank"`
	PP                     float64       `json:"pp"`
	RankedScore            int64         `json:"ranked_score"`
	HitAccuracy            float64       `json:"hit_accuracy"`
	PlayCount              int           `json:"play_count"`
	PlayTime               int           `json:"play_time"`
	TotalScore             int64         `json:"total_score"`
	TotalHits              int64         `json:"total_hits"`
	MaximumCombo           int           `json:"maximum_combo"`
	ReplaysWatchedByOthers int           `json:"replays_watched_by_others"`
	IsRanked               bool          `json:"is_ranked"`
	GradeCounts            GradeCounts   `json:"grade_counts"`
	Variants               []UserVariant `json:"variants"`
}

type DateOnly time.Time

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}

	*d = DateOnly(t)
	return nil
}

type UserLevel struct {
	Current  int `json:"current"`
	Progress int `json:"progress"`
}

type GradeCounts struct {
	SS  int `json:"ss"`
	SSH int `json:"ssh"`
	S   int `json:"s"`
	SH  int `json:"sh"`
	A   int `json:"a"`
}

type UserVariant struct {
	Mode       string `json:"mode"`
	Variant    string `json:"variant"`
	GlobalRank *int   `json:"global_rank"`
}

type UserAchievement struct {
	AchievementID int       `json:"achievement_id"`
	AchievedAt    time.Time `json:"achieved_at"`
}

type UsersResponse struct {
	Users []UserExtended `json:"users"`
}

type UserCache struct {
	m  map[int]struct{}
	mu sync.RWMutex
}

func (c *UserCache) Add(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[id] = struct{}{}
}

func (c *UserCache) Delete(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, id)
}

func (c *UserCache) Exists(id int) bool {
	c.mu.RLock()
	_, exists := c.m[id]
	c.mu.RUnlock()
	return exists
}

var userCache = &UserCache{
	m: make(map[int]struct{}),
}

func (u *UserExtended) Create() error {
	_, err := DB.Exec(context.Background(), `
    INSERT INTO users (
        user_id,
        username,
        username_safe,
        country
    ) VALUES (
        $1, $2, $3, $4
    ) ON CONFLICT (user_id) DO NOTHING
	`,
		u.ID,
		u.Username,
		u.Safename(),
		u.CountryCode,
	)

	return err
}

func (u *UserExtended) Update() error {
	_, err := DB.Exec(context.Background(), `
    UPDATE users SET username = $1, username_safe = $2, country = $3, restricted = 0 WHERE user_id = $4`,
		u.Username,
		u.Safename(),
		u.CountryCode,
		u.ID,
	)

	return err
}

func (u *UserExtended) GetRecent(mode string) ([]Score, error) {
	body, err := Fetch(fmt.Sprintf("/users/%d/scores/recent?mode=%s&include_fails=%d&limit=100", u.ID, mode, includeFailed))
	if err != nil {
		return nil, err
	}

	var data []Score

	err = json.Unmarshal(body, &data)
	return data, err
}

func (u *UserExtended) UpdateScores(mode string) error {
	data, err := u.GetRecent(mode)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	for _, score := range data {
		if err := score.Insert(); err != nil {
			return err
		}
		scoreCache.Set(score.ID, struct{}{}, time.Until(score.EndedAt.Add(24*time.Hour)))
		score.Beatmap.Insert(score.Beatmapset)
	}
	return nil
}

func (u *UserExtended) Restrict() error {
	row := DB.QueryRow(context.Background(), `
    UPDATE users SET restricted = 1 WHERE user_id = $1 RETURNING username`,
		u.ID,
	)

	userCount--

	err := row.Scan(&u.Username)
	log.Printf("%s (%d) just got restricted!", u.Username, u.ID)

	return err
}

func (u *UserStatistics) UpdateHistory(id int, mode int) error {
	global := 999999999
	if u.GlobalRank != nil {
		global = *u.GlobalRank
	}

	country := 999999999
	if u.CountryRank != nil {
		country = *u.CountryRank
	}

	_, err := DB.Exec(context.Background(), `
	INSERT INTO stats (
		user_id, mode, global, country, pp, accuracy,
		playcount, playtime, score, hits, level,
		progress, replays_watched
	)
	VALUES (
		$1, $2, $3, $4, $5, $6,
        $7, $8, $9, $10, $11,
        $12, $13
	)
	ON CONFLICT (user_id, mode, day)
	DO UPDATE SET
		global           = EXCLUDED.global,
		country          = EXCLUDED.country,
		pp               = EXCLUDED.pp,
		accuracy         = EXCLUDED.accuracy,
		playcount        = EXCLUDED.playcount,
		playtime         = EXCLUDED.playtime,
		score            = EXCLUDED.score,
		hits             = EXCLUDED.hits,
		level            = EXCLUDED.level,
		progress         = EXCLUDED.progress,
		replays_watched  = EXCLUDED.replays_watched
	WHERE
			stats.global          IS DISTINCT FROM EXCLUDED.global
		OR  stats.country         IS DISTINCT FROM EXCLUDED.country
		OR  stats.pp              IS DISTINCT FROM EXCLUDED.pp
		OR  stats.accuracy        IS DISTINCT FROM EXCLUDED.accuracy
		OR  stats.playcount       IS DISTINCT FROM EXCLUDED.playcount
		OR  stats.playtime        IS DISTINCT FROM EXCLUDED.playtime
		OR  stats.score           IS DISTINCT FROM EXCLUDED.score
		OR  stats.hits            IS DISTINCT FROM EXCLUDED.hits
		OR  stats.level           IS DISTINCT FROM EXCLUDED.level
		OR  stats.progress        IS DISTINCT FROM EXCLUDED.progress
		OR  stats.replays_watched IS DISTINCT FROM EXCLUDED.replays_watched;
    `,
		id,
		mode,
		global,
		country,
		u.PP,
		u.HitAccuracy,
		u.PlayCount,
		u.PlayTime,
		u.TotalScore,
		u.TotalHits,
		u.Level.Current,
		u.MaximumCombo,
		u.ReplaysWatchedByOthers,
	)

	return err
}

func (u *UserExtended) UpdateBase() error {
	_, err := DB.Exec(context.Background(), `
	INSERT INTO stats_base (
		user_id,
		badges,
		followers,
		achievements
	)
	VALUES (
		$1, $2, $3, $4
	)
	ON CONFLICT (user_id, day)
	DO UPDATE SET
		badges        = EXCLUDED.badges,
		banners       = EXCLUDED.banners,
		followers     = EXCLUDED.followers,
		achievements  = EXCLUDED.achievements
	WHERE
			stats_base.badges       IS DISTINCT FROM EXCLUDED.badges
		OR  stats_base.banners      IS DISTINCT FROM EXCLUDED.banners
		OR  stats_base.followers    IS DISTINCT FROM EXCLUDED.followers
		OR  stats_base.achievements IS DISTINCT FROM EXCLUDED.achievements;
    `,
		u.ID,
		len(u.Badges),
		u.FollowerCount,
		len(u.UserAchievements),
	)

	return err
}

func (u *UserExtended) Fetch(mode int) error {
	body, err := Fetch(fmt.Sprintf("/users/%d?mode=%d", u.ID, mode))
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &u)
}

func (u *UserExtended) Safename() string {
	return strings.ReplaceAll(strings.ToLower(u.Username), " ", "_")
}

func loadUsers() {
	rows, err := DB.Query(context.Background(),
		"SELECT user_id FROM users WHERE restricted = 0;",
	)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return
		}
		userCache.Add(id)
		userCount++
	}
	log.Printf("Loaded %d users", userCount)
}

func loadQueue() {
	rows, err := DB.Query(context.Background(), `
	SELECT
		t.user_id,
		t.mode
	FROM (
		SELECT DISTINCT ON (s.user_id, s.mode)
			s.user_id,
			s.mode,
			s.time
		FROM scores s
		JOIN users u ON u.user_id = s.user_id
		WHERE u.restricted = 0
		AND s.time > u.last_update
		ORDER BY
			s.user_id,
			s.mode,
			s.time ASC
	) t
	ORDER BY t.time ASC;
    `)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var c atomic.Int64

	for rows.Next() {
		var id, mode int
		if err := rows.Scan(&id, &mode); err != nil {
			return
		}
		c.Add(1)
		go userUpdater.Queue(id, uint8(mode), true)
	}

	log.Printf("Queued %d updates\n", c.Load())
}

func updateUser(id int, modes uint8) error {
	user := UserExtended{ID: id}

	for i := 0; i < 4; i++ {
		if modes&(1<<i) != 0 {
			if err := user.Fetch(i); err != nil {
				if err == ErrNotFound {
					user.Restrict()
					return nil
				}
				os.WriteFile("date.error", fmt.Appendf(nil, "%d: %s", id, err.Error()), 0644)
				return err
			}

			user.UpdateScores(ModeStr(i))
			user.Statistics.UpdateHistory(id, i)
			log.Printf("Updated %s (%d) on Mode %d", user.Username, user.ID, i)
		}
	}
	statsCount++
	user.Update()
	user.UpdateBase()
	return nil
}

func ModeStr(mode int) string {
	switch mode {
	case 0:
		return "osu"
	case 1:
		return "taiko"
	case 2:
		return "fruits"
	case 3:
		return "mania"
	default:
		return "osu"
	}
}
