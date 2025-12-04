package main

import (
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

	Statistics       *UserStatistics   `json:"statistics"`
	SupportLevel     int               `json:"support_level"`
	UserAchievements []UserAchievement `json:"user_achievements"`
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

type Update struct {
	Standard bool
	Taiko    bool
	Catch    bool
	Mania    bool
}

// type UserCache struct {
// 	m  map[int]Update
// 	mu sync.RWMutex
// }

// var UserCache = &UserCache{
// 	m: make(map[int]Update),
// }

func fetchUsers() {

}

func (u *UserExtended) Create() {

}
