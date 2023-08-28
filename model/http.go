package model

import "go-pinterest/db"

const (
	HttpSuccess = 0
	HttpFail    = 1
)

type HttpResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type SignInInfo struct {
	DeviceId   string `json:"device_id"`
	UserName   string `json:"user_name"`
	Password   string `json:"password"`
	Provider   string `json:"provider"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	BirthDay   string `json:"birth_day"`
	Gender     string `json:"gender"`
	AppVersion string `json:"app_version"`
}

type MatchResultUpload struct {
	UserId   int         `json:"user_id"`
	AudioId  int         `json:"audio_id"`
	Score    int         `json:"score"`
	Accuracy float32     `json:"accuracy"`
	PoseTime float32     `json:"pose_time"`
	VideoMd5 string      `json:"video_md5"`
	CoverMd5 string      `json:"cover_md5"`
	PlayInfo interface{} `json:"play_info"`
	DeviceId string      `json:"device_id"`
	PlayMode string      `json:"play_mode"`
}
type SignedUserInfo struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	BirthDay    string `json:"birth_day"`
	Gender      string `json:"gender"`
	AccessToken string `json:"access_token"`
	ExpireTime  int64  `json:"expire_time"`
	AccuracyAvg string `json:"accuracy_avg"`
	TimeAvg     string `json:"time_avg"`
	TotalScore  string `json:"total_score"`
	Rank        int    `json:"rank"`
}

type BasicUserInfo struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	BirthDay    string `json:"birth_day"`
	Gender      string `json:"gender"`
	AccuracyAvg string `json:"accuracy_avg"`
	TimeAvg     string `json:"time_avg"`
	TotalScore  string `json:"total_score"`
	Rank        int    `json:"rank"`
}
type AccessToken struct {
	User   string `json:"user"`
	Expire int64  `json:"time"`
}

type SongBasicInfo struct {
	UnLock           bool          `json:"un_lock"`
	TiktokVideoId    string        `json:"tiktok_video_id"`
	AudioCover       string        `json:"audio_cover"`
	AudioId          string        `json:"audio_id"`
	AudioTitle       string        `json:"audio_title"`
	AudioUrl         string        `json:"audio_url"`
	TiktokAuthorId   string        `json:"tiktok_author_id"`
	AuthorName       string        `json:"author_name"`
	AuthorUserName   string        `json:"author_user_name"`
	VideoAudio       string        `json:"video_audio"`
	VideoCover       string        `json:"video_cover"`
	VideoDescription string        `json:"video_description"`
	VideoUrl         string        `json:"video_url"`
	VideoMd5         string        `json:"video_md5"`
	VideoDuration    string        `json:"video_duration"`
	TotalMatch       int           `json:"total_match"`
	TotalPlayer      int           `json:"total_player"`
	BestResult       ScoreResponse `json:"best_result"`
}

type YoutubeVideoInfo struct {
	Id        string `json:"id"`
	ShortUrl  string `json:"short_url"`
	Thumbnail string `json:"thumbnail"`
	Url       string `json:"url"`
	MatchId   int    `json:"match_id"`
}

type SongDetailInfo struct {
	Info   SongBasicInfo `json:"info"`
	Frames interface{}   `json:"frames"`
}

type ScoreResponse struct {
	Match   *db.MatchResult   `json:"match"`
	Youtube *YoutubeVideoInfo `json:"youtube"`
	User    *BasicUserInfo    `json:"user"`
}

type MessageResponse struct {
	Id           int      `json:"id"`
	Text         string   `json:"text"`
	Type         int      `json:"type"`
	ObjectId     int      `json:"object_id"`
	Sender       *db.User `json:"sender"`
	ReceivedTime int64    `json:"received_time"`
}

type ListDataResponse struct {
	Data       []interface{} `json:"data"`
	NextOffset string        `json:"next_offset"`
}
