package db

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Gender   string `json:"gender"`
	BirthDay string `json:"birth_day"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
	Provider string `json:"provider"`
}

type UserAndStat struct {
	UserId      int     `json:"id"`
	TotalGame   int     `json:"total_game"`
	TimeAvg     float32 `json:"time_avg"`
	AccuracyAvg float32 `json:"accuracy_avg"`
}

type UserAndDevice struct {
	UserId     int    `json:"user_id"`
	DeviceId   string `json:"device_id"`
	AppVersion string `json:"app_version"`
	LastAccess int64  `json:"last_access"`
}

type MatchResult struct {
	Id           int     `json:"id"`
	UserId       int     `json:"user_id"`
	AudioId      int     `json:"audio_id"`
	Video        string  `json:"video"`
	Cover        string  `json:"cover"`
	VideoMd5     string  `json:"video_md5"`
	CoverMd5     string  `json:"cover_md5"`
	Score        int     `json:"score"`
	Accuracy     float32 `json:"accuracy"`
	PoseTime     float32 `json:"pose_time"`
	PlayInfo     string  `json:"play_info"`
	ReceivedTime int64   `json:"received_time"`
	PlayMode     string  `json:"play_mode"`
}

type ListMatch struct {
	MatchId int `json:"match_id"`
	UserId  int `json:"user_id"`
	SongId  int `json:"song_id"`
	Score   int `json:"score"`
	Time    int `json:"time"`
}

type MatchAndYoutube struct {
	MatchId   int    `json:"match_id"`
	YoutubeId string `json:"youtube_id"`
	Thumbnail string `json:"thumbnail"`
}
