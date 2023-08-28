package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-pinterest/db"
	"go-pinterest/model"
	"io"
	"net/http"
	"strings"
)

var MapSongs = map[string]model.SongDetailInfo{}

func InitListAudio() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error when reload list new audio ", r)
		}
	}()
	response, err := http.DefaultClient.Get("https://mangaverse.skymeta.pro/meme/pinter_meme_v3/pose-dance/song-dance-list")
	if err != nil {
		println("error when init audios", err.Error())
	}
	text, err := io.ReadAll(response.Body)
	if err != nil {
		println("error when read all", err.Error())
	}
	body := map[string]interface{}{}
	json.Unmarshal(text, &body)
	listData := body["data"].([]interface{})

	var NewMapSongs = map[string]model.SongDetailInfo{}
	for _, dataInterface := range listData {
		data := dataInterface.(map[string]interface{})
		audioInfo := data["info"].(map[string]interface{})
		videoDuration := audioInfo["video_duration"].(float64)
		videoDurationText := fmt.Sprintf("%f", videoDuration)
		songInfo := model.SongBasicInfo{
			UnLock:           true,
			TiktokVideoId:    audioInfo["Id"].(string),
			AudioCover:       audioInfo["audio_cover"].(string),
			AudioId:          audioInfo["audio_id"].(string),
			AudioTitle:       audioInfo["audio_title"].(string),
			AudioUrl:         audioInfo["audio_url"].(string),
			TiktokAuthorId:   audioInfo["author_id"].(string),
			AuthorName:       audioInfo["author_name"].(string),
			AuthorUserName:   audioInfo["author_user_name"].(string),
			VideoCover:       audioInfo["cover"].(string),
			VideoDescription: audioInfo["description"].(string),
			VideoUrl:         audioInfo["url"].(string),
			VideoMd5:         audioInfo["video_md5"].(string),
			VideoAudio:       audioInfo["video_audio"].(string),
			VideoDuration:    videoDurationText,
		}
		frames := data["frames"]
		songDetail := model.SongDetailInfo{Info: songInfo, Frames: frames}
		NewMapSongs[songDetail.Info.AudioId] = songDetail
	}
	if len(NewMapSongs) > 0 {
		MapSongs = NewMapSongs
	}
}

func GetAudioInfo(c *gin.Context) {
	audioId := strings.ToLower(c.Param("audio_id"))
	song := MapSongs[audioId]
	var err error
	song.Info.TotalMatch, err = db.GetTotalMatchResultByAudioId(song.Info.AudioId)
	if err != nil {
		println("err when get total match by audio id ", song.Info.AudioId, " err ", err.Error())
	}
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: song})
	c.Data(200, "text/html; charset=UTF-8", data)
}
