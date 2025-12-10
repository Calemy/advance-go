package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	discordwebhook "github.com/bensch777/discord-webhook-golang"
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
	StartDate time.Time `json:"start_date"`
	Count     int       `json:"count"`
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

type Update struct {
	Standard bool
	Taiko    bool
	Catch    bool
	Mania    bool
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

var trackCache = &UserCache{
	m: make(map[int]struct{}),
}

var playedCache = &UserCache{
	m: make(map[int]struct{}),
}

var userBatcher = NewBatcher(50, time.Minute, fetchUsers)
var trackBatcher = NewBatcher(50, time.Minute, fetchUsers)

func updateEmptyUsers() {
	rows, err := DB.Query(context.Background(), "SELECT user_id FROM users_go WHERE username = '' ORDER BY added ASC LIMIT 50")
	if err != nil {
		return
	}

	defer rows.Close()

	queued := 0

	for rows.Next() {
		var id int

		if trackCache.Exists(id) {
			continue
		}

		if err := rows.Scan(&id); err != nil {
			log.Fatalln(err)
			return
		}

		trackBatcher.Add(id)
		trackCache.Add(id)
		queued++
	}

	if queued > 0 {
		log.Printf("Queued %d users to start tracking", queued)
	}
}

func updateUsers() {
	for user := range playedCache.m {
		userBatcher.Add(user)
	}

	log.Printf("Queued %d users for a scheduled update", len(playedCache.m))
}

func fetchUsers(ids []int) {
	idset := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idset[id] = struct{}{}
		playedCache.Delete(id)
	}

	body, err := Fetch(fmt.Sprintf("/users?include_variant_statistics=true&ids[]=%s", JoinInts(ids, "&ids[]=")))
	if err != nil {
		return
	}

	var resp UsersResponse

	if err := json.Unmarshal(body, &resp); err != nil {
		return
	}

	var wg sync.WaitGroup

	for _, user := range resp.Users {
		if _, ok := idset[user.ID]; !ok {
			user.Restrict()
			continue
		}

		wg.Add(1)
		go func(u UserExtended) {
			defer wg.Done()
			u.Update()
			var innerWg sync.WaitGroup
			innerWg.Add(len(*u.StatisticsRulesets)) //Incase peppy decides to bug out
			for mode, stats := range *u.StatisticsRulesets {
				go func(s UserStatistics, mode string) {
					defer innerWg.Done()
					if !s.IsRanked || s.PP == 0 {
						return
					}

					s.UpdateHistory(user.ID, ModeInt(mode))
				}(stats, mode)
			}
			innerWg.Wait()
		}(user)
	}
	wg.Wait()
}

func (u *UserExtended) Create() error {
	_, err := DB.Exec(context.Background(), `
    INSERT INTO users_go (
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
    UPDATE users_go SET username = $1, username_safe = $2, country = $3 WHERE user_id = $4`,
		u.Username,
		u.Safename(),
		u.CountryCode,
		u.ID,
	)

	trackCache.Delete(u.ID)

	userCount++

	embed := discordwebhook.Embed{
		Title:       fmt.Sprintf("%s (%d) is now tracked!", u.Username, u.ID),
		Description: "Welcome to the community!",
		Color:       0x86DC3D,
		Timestamp:   time.Now(),
		Thumbnail: discordwebhook.Thumbnail{
			Url: fmt.Sprintf("https://a.ppy.sh/%d", u.ID),
		},
		Footer: discordwebhook.Footer{
			Text: fmt.Sprintf("Users tracked: %d", userCount),
		},
	}

	hook := discordwebhook.Hook{
		Username:   "Advance",
		Avatar_url: "https://a.ppy.sh/9527931",
		Embeds:     []discordwebhook.Embed{embed},
	}

	go func(hook discordwebhook.Hook) {
		webhookQueue <- hook
	}(hook)

	return err
}

func (u *UserExtended) Restrict() error {
	row := DB.QueryRow(context.Background(), `
    UPDATE users_go SET restricted = 1 WHERE user_id = $1 RETURNING username`,
		u.ID,
	)

	userCount--

	err := row.Scan(&u.Username)
	log.Printf("%s (%d) just got restricted!", u.Username, u.ID)

	embed := discordwebhook.Embed{
		Title:       fmt.Sprintf("%s (%d) just got restricted!", u.Username, u.ID),
		Description: "We can only hope they didn't cheat",
		Color:       0xD2042D,
		Timestamp:   time.Now(),
		Thumbnail: discordwebhook.Thumbnail{
			Url: fmt.Sprintf("https://a.ppy.sh/%d", u.ID),
		},
		Footer: discordwebhook.Footer{
			Text: fmt.Sprintf("Users tracked: %d", userCount),
		},
	}

	hook := discordwebhook.Hook{
		Username:   "Advance",
		Avatar_url: "https://a.ppy.sh/9527931",
		Embeds:     []discordwebhook.Embed{embed},
	}

	go func(hook discordwebhook.Hook) {
		webhookQueue <- hook
	}(hook)

	return err
}

func (u *UserStatistics) UpdateHistory(userID int, mode int) error {
	global := 999999999
	if u.GlobalRank != nil {
		global = *u.GlobalRank
	}

	country := 999999999
	if u.CountryRank != nil {
		country = *u.CountryRank
	}

	_, err := DB.Exec(context.Background(), `
        MERGE INTO stats_go AS t
        USING (
            SELECT
                $1::int       AS user_id,
                $13::int      AS mode,
                $2::int       AS global,
                $3::int       AS country,
                $4::float8    AS pp,
                $5::float8    AS accuracy,
                $6::int       AS playcount,
                $7::int       AS playtime,
                $8::bigint    AS score,
                $9::bigint    AS hits,
                $10::int      AS level,
                $11::int      AS progress,
                $12::int      AS replays_watched
        ) AS incoming
        ON (
            t.user_id = incoming.user_id
            AND t.mode = incoming.mode
            AND t.time >= date_trunc('day', NOW())
        )
        WHEN MATCHED AND (
                incoming.global          IS DISTINCT FROM t.global
            OR  incoming.country         IS DISTINCT FROM t.country
            OR  incoming.pp              IS DISTINCT FROM t.pp
            OR  incoming.accuracy        IS DISTINCT FROM t.accuracy
            OR  incoming.playcount       IS DISTINCT FROM t.playcount
            OR  incoming.playtime        IS DISTINCT FROM t.playtime
            OR  incoming.score           IS DISTINCT FROM t.score
            OR  incoming.hits            IS DISTINCT FROM t.hits
            OR  incoming.level           IS DISTINCT FROM t.level
            OR  incoming.progress        IS DISTINCT FROM t.progress
            OR  incoming.replays_watched IS DISTINCT FROM t.replays_watched
        )
        THEN UPDATE SET
            global           = incoming.global,
            country          = incoming.country,
            pp               = incoming.pp,
            accuracy         = incoming.accuracy,
            playcount        = incoming.playcount,
            playtime         = incoming.playtime,
            score            = incoming.score,
            hits             = incoming.hits,
            level            = incoming.level,
            progress         = incoming.progress,
            replays_watched  = incoming.replays_watched,
            time             = NOW()
        WHEN NOT MATCHED THEN
            INSERT (
                user_id, mode, global, country, pp, accuracy,
                playcount, playtime, score, hits, level,
                progress, replays_watched, time
            ) VALUES (
                incoming.user_id, incoming.mode,
                incoming.global, incoming.country, incoming.pp, incoming.accuracy,
                incoming.playcount, incoming.playtime, incoming.score, incoming.hits, incoming.level,
                incoming.progress, incoming.replays_watched,
                NOW()
            )
    `,
		userID,
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
		mode,
	)

	return err
}

func (u *UserExtended) Fetch() error {
	body, err := Fetch(fmt.Sprintf("/users/%d", u.ID))
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &u)
}

func (u *UserExtended) Safename() string {
	return strings.ReplaceAll(strings.ToLower(u.Username), " ", "_")
}

func loadUsers() {
	rows, err := DB.Query(context.Background(), "SELECT user_id FROM users_go")
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Fatalln(err)
			return
		}
		userCache.Add(id)
	}
}

func ModeInt(mode string) int {
	switch mode {
	case "osu":
		return 0
	case "taiko":
		return 1
	case "fruits", "ctb":
		return 2
	case "mania":
		return 3
	default:
		return -1 // invalid
	}
}
