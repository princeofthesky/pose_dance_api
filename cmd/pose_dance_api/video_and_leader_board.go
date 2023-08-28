package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-pinterest/date"
	"go-pinterest/db"
	"go-pinterest/model"
	"io"
	"math"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"
)

var USER_NOT_LOGIN = db.User{}

func InitUserNotLogin() {
	USER_NOT_LOGIN, _ = db.GetUserById(-1)
	println("init user not login ", USER_NOT_LOGIN.Id, USER_NOT_LOGIN.Name, USER_NOT_LOGIN.Avatar)
}

func UploadVideoMatchResultByMultipart(c *gin.Context) {
	matchIdText := strings.ToLower(c.Param("match_id"))
	matchId, err := strconv.Atoi(matchIdText)
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when parser match id ", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}

	forms, err := c.MultipartForm()
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "mutil part form not found", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	var videoForm *multipart.FileHeader = nil
	if len(forms.File["video"]) > 0 {
		videoForm = forms.File["video"][0]
	} else {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "video not found", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}

	var coverForm *multipart.FileHeader = nil
	if len(forms.File["cover"]) > 0 {
		coverForm = forms.File["cover"][0]
	} else {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "cover not found", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}

	matchResult, _ := db.GetMatchResultById(matchId)
	if matchResult.Id == 0 {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "match Id not found : " + matchIdText, Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	if len(matchResult.Video) > 0 {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpSuccess, Message: "SUCCESS", Data: matchResult})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	videoMd5Got := ""
	coverMd5Got := ""
	if videoForm != nil {
		videoFile, _ := videoForm.Open()
		videoData, _ := io.ReadAll(videoFile)
		coverFile, _ := coverForm.Open()
		coverData, _ := io.ReadAll(coverFile)
		if len(coverData) > 0 && len(videoData) > 0 {
			videoMd5Byte := md5.Sum(videoData)
			videoMd5Got = strings.ToLower(hex.EncodeToString(videoMd5Byte[:]))
			if strings.Compare(videoMd5Got, matchResult.VideoMd5) != 0 {
				reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when compare  video md5 : want  " + matchResult.VideoMd5 + " , received : " + videoMd5Got, Data: nil})
				c.Data(200, "text/html; charset=UTF-8", reposeData)
				return
			}

			coverMd5Byte := md5.Sum(coverData)
			coverMd5Got = strings.ToLower(hex.EncodeToString(coverMd5Byte[:]))

			if strings.Compare(coverMd5Got, matchResult.CoverMd5) != 0 {
				reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when compare  cover md5 : want  " + matchResult.CoverMd5 + " , received : " + coverMd5Got, Data: nil})
				c.Data(200, "text/html; charset=UTF-8", reposeData)
				return
			}
			folderToday := date.GetFolderDaily(matchResult.ReceivedTime)
			videoFolder := *videoDir + folderToday
			err = os.MkdirAll(videoFolder, os.ModePerm)
			if err != nil {
				println("err when make folder ", videoFolder, "err", err.Error())
			}
			videoFileName := videoFolder + "/" + videoMd5Got + ".mp4"
			println("video File Name ", videoFileName, "match Id", matchId)
			f, err := os.Create(videoFileName)
			f.Write(videoData)
			f.Sync()
			f.Close()

			coverFolder := *videoCoverDir + folderToday
			err = os.MkdirAll(coverFolder, os.ModePerm)
			if err != nil {
				println("err when make folder ", coverFolder, "err", err.Error())
			}
			coverFileName := coverFolder + "/" + coverMd5Got + ".png"
			println("cover File Name ", coverFileName)
			f, err = os.Create(coverFileName)
			f.Write(coverData)
			f.Sync()
			f.Close()

			matchResult.VideoMd5 = videoMd5Got
			matchResult.CoverMd5 = coverMd5Got
			matchResult.Video = "https://mangaverse.skymeta.pro/meme_videos/" + folderToday + "/" + videoMd5Got + ".mp4"
			matchResult.Cover = "https://mangaverse.skymeta.pro/meme_covers/" + folderToday + "/" + coverMd5Got + ".png"
			matchResult, err = db.UpdateVideoMatchResult(matchResult)
			if err != nil {
				println("error when update video", err.Error())
			}
		}
	}
	reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpSuccess, Message: "SUCCESS", Data: matchResult})
	println("reponse => ", string(reposeData))
	c.Data(200, "text/html; charset=UTF-8", reposeData)
}

func UploadMatchResult(c *gin.Context) {
	request, _ := io.ReadAll(c.Request.Body)
	uploadedResult := model.MatchResultUpload{}
	err := json.Unmarshal(request, &uploadedResult)
	if err != nil {
		fmt.Println("Error when parse json upload match result", err)
	}
	println("Try upload match result with body ", string(request))

	user, _ := db.GetUserById(uploadedResult.UserId)
	if user.Id <= 0 {
		user = USER_NOT_LOGIN
	}
	uploadedResult.VideoMd5 = strings.TrimSpace(uploadedResult.VideoMd5)
	if len(uploadedResult.VideoMd5) == 0 {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "video md5 not found : " + uploadedResult.VideoMd5, Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}

	if _, exit := MapSongs[strconv.Itoa(uploadedResult.AudioId)]; !exit {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "audio id not exit : " + strconv.Itoa(uploadedResult.AudioId), Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	matchResult, err := db.GetMatchResultByVideoMd5(uploadedResult.VideoMd5)
	if err == nil && matchResult.Id > 0 {
		userStat, _ := db.GetUserAndStatById(uploadedResult.UserId)
		totalScore, rank, _ := db.GetRankByUser(uploadedResult.UserId)
		scoreResponse := model.ScoreResponse{
			Match: &matchResult,
			User: &model.BasicUserInfo{
				Id:          matchResult.Id,
				AccuracyAvg: fmt.Sprintf("%.10f", userStat.AccuracyAvg),
				TimeAvg:     fmt.Sprintf("%.10f", userStat.TimeAvg),
				TotalScore:  fmt.Sprintf("%.10f", totalScore),
				Rank:        rank,
			},
		}
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpSuccess, Message: "SUCCESS", Data: scoreResponse})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}

	matchResult.UserId = user.Id
	matchResult.AudioId = uploadedResult.AudioId
	playInfo, _ := json.Marshal(uploadedResult.PlayInfo)
	matchResult.PlayInfo = string(playInfo)
	matchResult.ReceivedTime = time.Now().Unix()
	matchResult.Score = uploadedResult.Score
	matchResult.Accuracy = uploadedResult.Accuracy
	matchResult.VideoMd5 = uploadedResult.VideoMd5
	matchResult.CoverMd5 = uploadedResult.CoverMd5
	matchResult.PoseTime = uploadedResult.PoseTime
	matchResult.PlayMode = uploadedResult.PlayMode
	matchResult, err = db.InsertMatchResult(matchResult)
	if err != nil {
		println("error when insert video", err.Error())
	}
	if user.Id > 0 {
		userStat, err := db.GetUserAndStatById(user.Id)
		if userStat.UserId == 0 {
			userStat.UserId = user.Id
		}
		if userStat.TotalGame == 0 {
			userStat.TotalGame = 1
			userStat.TimeAvg = float32(uploadedResult.PoseTime)
			userStat.AccuracyAvg = float32(uploadedResult.Accuracy)
		} else {
			userStat.TimeAvg = (float32(uploadedResult.PoseTime) + userStat.TimeAvg*float32(userStat.TotalGame)) / float32(userStat.TotalGame+1)
			userStat.AccuracyAvg = (float32(uploadedResult.Accuracy) + userStat.AccuracyAvg*float32(userStat.TotalGame)) / float32(userStat.TotalGame+1)
			userStat.TotalGame++
		}
		userStat, err = db.InsertUserAndStatInfo(userStat)
		if err != nil {
			println("error when insert user and stat :", err.Error())
		}
		totalScore, rank, err := db.UpdateTotalScoreByUser(user.Id, uploadedResult.Score)
		if err != nil {
			println("error when update total score & rank :", err.Error())
		}
		scoreResponse := model.ScoreResponse{
			Match: &matchResult,
			User: &model.BasicUserInfo{
				Id:          user.Id,
				AccuracyAvg: fmt.Sprintf("%.10f", userStat.AccuracyAvg),
				TimeAvg:     fmt.Sprintf("%.10f", userStat.TimeAvg),
				TotalScore:  fmt.Sprintf("%.10f", totalScore),
				Rank:        rank,
			},
		}
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpSuccess, Message: "SUCCESS", Data: scoreResponse})
		println("reponse => ", string(reposeData))
		c.Data(200, "text/html; charset=UTF-8", reposeData)
	} else {
		scoreResponse := model.ScoreResponse{
			Match: &matchResult,
			User: &model.BasicUserInfo{
				Id:          uploadedResult.UserId,
				AccuracyAvg: "0",
				TimeAvg:     "0",
				TotalScore:  "0",
				Rank:        -1,
			},
		}
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpSuccess, Message: "SUCCESS", Data: scoreResponse})
		println("reponse => ", string(reposeData))
		c.Data(200, "text/html; charset=UTF-8", reposeData)
	}

}

func GetMatchResultInfo(c *gin.Context) {
	matchIdText := strings.ToLower(c.Param("match_id"))
	matchId, err := strconv.Atoi(matchIdText)
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when parser video id ", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	match, err := db.GetMatchResultById(matchId)
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when get video info from database with video id = " + matchIdText, Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	responseData, _ := json.Marshal(model.HttpResponse{
		Code:    model.HttpSuccess,
		Message: "SUCCESS",
		Data:    match,
	})
	c.Data(200, "text/html; charset=UTF-8", responseData)
}

func GetListMatchResultInHomePage(c *gin.Context) {
	offSetText, exit := c.GetQuery("offset")
	if !exit {
		offSetText = "0"
	}
	lengthText, exit := c.GetQuery("length")
	if !exit {
		lengthText = "20"
	}
	offSet, err := strconv.Atoi(offSetText)
	if err != nil {
		offSet = 0
	}
	length, err := strconv.Atoi(lengthText)
	if err != nil {
		length = 20
	}
	if length > 20 {
		length = 20
	}
	listScores := model.ListDataResponse{}
	listScores.Data = []interface{}{}
	for i := offSet; i < offSet+length; i++ {
		if i >= len(ListYoutubeMatchResult) {
			break
		}
		scoreResponse := model.ScoreResponse{}
		match := ListYoutubeMatchResult[i]
		scoreResponse.Youtube = MapYoutubeVideos[match.Id]
		scoreResponse.Match = &match
		user, err := db.GetUserById(match.UserId)
		if err != nil {
			println("error when get user by id ", match.UserId, err)
			continue
		}
		scoreResponse.User = &model.BasicUserInfo{
			Id:     user.Id,
			Name:   user.Name,
			Avatar: user.Avatar,
		}
		listScores.Data = append(listScores.Data, scoreResponse)
	}
	if len(ListYoutubeMatchResult) <= offSet+length {
		listScores.NextOffset = "-1"
	} else {
		listScores.NextOffset = strconv.Itoa(offSet + length)
	}
	responseData, _ := json.Marshal(model.HttpResponse{
		Code:    model.HttpSuccess,
		Message: "",
		Data:    listScores,
	})
	c.Data(200, "text/html; charset=UTF-8", responseData)
}

func GetListMatchResult(c *gin.Context) {
	offSetText, exit := c.GetQuery("offset")
	if !exit {
		offSetText = "0"
	}
	lengthText, exit := c.GetQuery("length")
	if !exit {
		lengthText = "20"
	}
	offSet, err := strconv.Atoi(offSetText)
	if err != nil {
		offSet = 0
	}
	length, err := strconv.Atoi(lengthText)
	if err != nil {
		length = 20
	}
	if length > 20 {
		length = 20
	}
	if offSet == 0 {
		offSet = int(time.Now().Unix())
	}
	matchIds, err := db.GetListMatchResultId(offSet, length)
	if err != nil {
		println("error when get list new video ", err.Error())
	}
	listScores := model.ListDataResponse{}
	listScores.Data = []interface{}{}
	for i := 0; i < len(matchIds); i++ {
		scoreResponse := model.ScoreResponse{}
		match, err := db.GetMatchResultById(matchIds[i].Id)
		if err != nil {
			println("error when get video by id ", matchIds[i].Id, err)
			continue
		}
		scoreResponse.Youtube = MapYoutubeVideos[matchIds[i].Id]
		scoreResponse.Match = &match
		user, err := db.GetUserById(matchIds[i].UserId)
		if err != nil {
			println("error when get user by id ", matchIds[i].UserId, err)
			continue
		}
		scoreResponse.User = &model.BasicUserInfo{
			Id:     user.Id,
			Name:   user.Name,
			Avatar: user.Avatar,
		}
		listScores.Data = append(listScores.Data, scoreResponse)
	}
	if len(matchIds) < length {
		listScores.NextOffset = "-1"
	} else {
		size := len(listScores.Data)
		lastInfo := listScores.Data[size-1].(model.ScoreResponse)
		listScores.NextOffset = strconv.Itoa(lastInfo.Match.Score) + "." + strconv.FormatInt(lastInfo.Match.ReceivedTime, 10)
	}
	responseData, _ := json.Marshal(model.HttpResponse{
		Code:    model.HttpSuccess,
		Message: "",
		Data:    listScores,
	})
	c.Data(200, "text/html; charset=UTF-8", responseData)
}

func GetMatchResultByAudio(c *gin.Context) {
	audioId := strings.ToLower(c.Param("audio_id"))
	offSetText, _ := c.GetQuery("offset")
	lengthText, _ := c.GetQuery("length")
	queryType, _ := c.GetQuery("type")
	playMode, _ := c.GetQuery("play_mode")
	limitScore := math.MaxInt32
	limitTime := int64(0)
	var err error
	if offSetText != "0" {
		splitText := strings.Split(offSetText, ".")
		if len(splitText) == 2 {
			limitScore, _ = strconv.Atoi(splitText[0])
			limitTime, _ = strconv.ParseInt(splitText[1], 10, 64)
		}
	}
	length, err := strconv.Atoi(lengthText)
	if err != nil {
		length = 20
	}
	if length > 20 {
		length = 20
	}
	var scoreInfos []db.MatchResult
	if queryType == "video" {
		scoreInfos, err = db.GetTopVideoMatchResultByAudioId(audioId,playMode, limitScore, limitTime, length)
	} else {
		scoreInfos, err = db.GetTopMatchResultByAudioId(audioId,playMode, limitScore, limitTime, length)
	}
	println("scoreInfos", len(scoreInfos))
	if err != nil {
		println("error when get list new video ", err.Error())
	}
	listScores := model.ListDataResponse{}
	listScores.Data = []interface{}{}
	for i := 0; i < len(scoreInfos); i++ {
		scoreReponse := model.ScoreResponse{}
		match, err := db.GetMatchResultById(scoreInfos[i].Id)
		if err != nil {
			println("error when get video by id ", scoreInfos[i].Id, err)
			continue
		}
		scoreReponse.Youtube = MapYoutubeVideos[scoreInfos[i].Id]
		scoreReponse.Match = &match
		user, err := db.GetUserById(scoreInfos[i].UserId)
		if err != nil {
			println("error when get user by id ", scoreInfos[i].UserId, err)
			continue
		}
		scoreReponse.User = &model.BasicUserInfo{
			Id:     user.Id,
			Name:   user.Name,
			Avatar: user.Avatar,
		}
		listScores.Data = append(listScores.Data, scoreReponse)
	}
	if len(scoreInfos) < length {
		listScores.NextOffset = "-1"
	} else {
		size := len(listScores.Data)
		lastInfo := listScores.Data[size-1].(model.ScoreResponse)
		listScores.NextOffset = strconv.Itoa(lastInfo.Match.Score) + "." + strconv.FormatInt(lastInfo.Match.ReceivedTime, 10)
	}
	responseData, _ := json.Marshal(model.HttpResponse{
		Code:    model.HttpSuccess,
		Message: "",
		Data:    listScores,
	})
	c.Data(200, "text/html; charset=UTF-8", responseData)
}

func GetMatchResultByUser(c *gin.Context) {
	userIdText := strings.ToLower(c.Param("user_id"))
	userId, err := strconv.Atoi(userIdText)

	offSetText, exit := c.GetQuery("offset")
	if !exit {
		offSetText = "0"
	}
	lengthText, exit := c.GetQuery("length")
	if !exit {
		lengthText = "20"
	}
	limitScore := math.MaxInt32
	limitTime := int64(0)
	if offSetText != "0" {
		splitText := strings.Split(offSetText, ".")
		if len(splitText) == 2 {
			limitScore, err = strconv.Atoi(splitText[0])
			if err != nil {
				limitScore = math.MaxInt32
			}
			limitTime, err = strconv.ParseInt(splitText[1], 10, 64)
			if err != nil {
				limitTime = time.Now().Unix()
			}
		}
	}
	length, err := strconv.Atoi(lengthText)
	if err != nil {
		length = 20
	}
	if length > 20 {
		length = 20
	}
	scoreInfos, err := db.GetTopMatchResultByUser(userId, limitScore, limitTime, length)
	if err != nil {
		println("error when get list new video ", err.Error())
	}
	listScores := model.ListDataResponse{}
	listScores.Data = []interface{}{}
	for i := 0; i < len(scoreInfos); i++ {
		scoreResponse := model.ScoreResponse{}
		match, err := db.GetMatchResultById(scoreInfos[i].Id)
		if err != nil {
			println("error when get video by id ", scoreInfos[i].Id, err)
			continue
		}
		scoreResponse.Youtube = MapYoutubeVideos[scoreInfos[i].Id]
		scoreResponse.Match = &match
		listScores.Data = append(listScores.Data, scoreResponse)
	}
	if len(scoreInfos) < length {
		listScores.NextOffset = "-1"
	} else {
		size := len(listScores.Data)
		lastInfo := listScores.Data[size-1].(model.ScoreResponse)
		listScores.NextOffset = strconv.Itoa(lastInfo.Match.Score) + "." + strconv.FormatInt(lastInfo.Match.ReceivedTime, 10)
	}
	responseData, _ := json.Marshal(model.HttpResponse{
		Code:    model.HttpSuccess,
		Message: "",
		Data:    listScores,
	})
	c.Data(200, "text/html; charset=UTF-8", responseData)
}
