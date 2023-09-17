package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-pinterest/db"
	"go-pinterest/model"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var MapSongs = map[string]model.SongDetailInfo{}
var ListSong []model.SongDetailInfo
var length = 20

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
	var NewListSongs = []model.SongDetailInfo{}
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
		var err error
		songDetail.Info.TotalMatch, err = db.GetTotalMatchResultByAudioId(songDetail.Info.AudioId)
		if err != nil {
			println("err when get total match by audio id ", songDetail.Info.AudioId, " err ", err.Error())
		}
		songDetail.Info.TotalPlayer, err = db.GetTotalPlayerByAudioId(songDetail.Info.AudioId)
		if err != nil {
			println("err when get total player by audio id ", songDetail.Info.AudioId, " err ", err.Error())
		}
		bestMatches, err := db.GetBestMatchResultByAudioId(songDetail.Info.AudioId, 1)
		if err != nil {
			println("err when get best result by audio id ", songDetail.Info.AudioId, " err ", err.Error())
		}
		if len(bestMatches) > 0 && bestMatches[0].Id > 0 {
			user, err := db.GetUserById(bestMatches[0].UserId)
			if err != nil {
				println("err when get user info by id ", bestMatches[0].UserId, " err ", err.Error())
			}
			songDetail.Info.BestResult = model.ScoreResponse{
				Youtube: MapYoutubeVideos[bestMatches[0].Id],
				Match:   &bestMatches[0],
				User: &model.BasicUserInfo{
					Id:     user.Id,
					Name:   user.Name,
					Avatar: user.Avatar,
				},
			}
		}
		NewListSongs = append(NewListSongs, songDetail)
		NewMapSongs[songDetail.Info.AudioId] = songDetail
	}
	if len(NewMapSongs) > 0 {
		MapSongs = NewMapSongs
		ListSong = NewListSongs
	}
}

func GetListAudios(c *gin.Context) {
	var allSongs []model.SongBasicInfo
	pageText, exit := c.GetQuery("page")
	if !exit {
		pageText = "0"
	}
	page, _ := strconv.Atoi(pageText)
	start := page * length
	end := start + length
	for i := start; i < end; i++ {
		if i >= len(ListSong) {
			break
		}
		song := ListSong[i]
		allSongs = append(allSongs, song.Info)
	}
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: allSongs})
	c.Data(200, "text/html; charset=UTF-8", data)
}

func GetAudioInfo(c *gin.Context) {
	audioId := strings.ToLower(c.Param("audio_id"))
	song := MapSongs[audioId]
	var err error
	song.Info.TotalMatch, err = db.GetTotalMatchResultByAudioId(song.Info.AudioId)
	if err != nil {
		println("err when get total match by audio id ", song.Info.AudioId, " err ", err.Error())
	}
	song.Info.TotalPlayer, err = db.GetTotalPlayerByAudioId(song.Info.AudioId)
	if err != nil {
		println("err when get total player by audio id ", song.Info.AudioId, " err ", err.Error())
	}
	bestMatchs, err := db.GetBestMatchResultByAudioId(song.Info.AudioId, 1)
	if err != nil {
		println("err when get best result by audio id ", song.Info.AudioId, " err ", err.Error())
	}
	if len(bestMatchs) > 0 {
		if bestMatchs[0].Id > 0 {
			user, err := db.GetUserById(bestMatchs[0].UserId)
			if err != nil {
				println("err when get user info by id ", bestMatchs[0].UserId, " err ", err.Error())
			}
			song.Info.BestResult = model.ScoreResponse{
				Youtube: MapYoutubeVideos[bestMatchs[0].Id],
				Match:   &bestMatchs[0],
				User: &model.BasicUserInfo{
					Id:     user.Id,
					Name:   user.Name,
					Avatar: user.Avatar,
				},
			}
		}
	}
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: song})
	c.Data(200, "text/html; charset=UTF-8", data)
}
