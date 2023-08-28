package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"go-pinterest/model"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultAvatar = "default.png"
const Path_Service_Name = "/pose_dance_api/v1_0"

var (
	httpPort      = flag.String("http_port", "9081", "http_port listen")
	conf          = flag.String("conf", "./pose_dance_api.toml", "config run file *.toml")
	c             = config.CrawlConfig{}
	videoDir      = flag.String("video_dir", "/meme_videos/", "video meme direction")
	videoCoverDir = flag.String("cover_dir", "/meme_covers/", "cover meme direction")
	avatarDir     = flag.String("avatar_dir", "/meme_avatars/", "avatar meme direction")
)

func main() {
	flag.Parse()
	configBytes, err := ioutil.ReadFile(*conf)
	if err != nil {
		fmt.Println("err when read config file ", err, "file ", *conf)
	}
	err = toml.Unmarshal(configBytes, &c)
	if err != nil {
		fmt.Println("err when pass toml file ", err)
	}
	text, err := json.Marshal(c)
	fmt.Println("Success read config from toml file ", string(text))
	err = db.Init(c)
	if err != nil {
		fmt.Println("err when connect postgres", err)
	}
	defer db.Close()

	r := gin.Default()
	go func() {
		for true {
			InitListAudio()
			InitListYoutubeVideo()
			time.Sleep(5 * time.Minute)
		}
	}()
	InitListAudio()
	InitUserNotLogin()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST(Path_Service_Name+"/sign_in", SignIn)
	r.GET(Path_Service_Name+"/users/:user_id/info", GetUserInfo)
	r.POST(Path_Service_Name+"/users/:user_id/update", UpdateUserInfo)
	r.POST(Path_Service_Name+"/users/:user_id/sign_out", SignOutUserInfo)
	r.POST(Path_Service_Name+"/users/:user_id/remove", RemoveUserInfo)
	r.POST(Path_Service_Name+"/users/:user_id/matches/:match_id/remove", RemoveMatchInfo)

	r.GET(Path_Service_Name+"/audios/list", GetListAudios)
	r.GET(Path_Service_Name+"/audios/:audio_id/info", GetAudioInfo)
	// upload match result
	r.POST(Path_Service_Name+"/match_results/:match_id/upload/video", UploadVideoMatchResultByMultipart)
	r.POST(Path_Service_Name+"/match_results/upload/score", UploadMatchResult)
	// leader board
	r.GET(Path_Service_Name+"/match_results", GetListMatchResult)
	r.GET(Path_Service_Name+"/match_results/home_page", GetListMatchResultInHomePage)
	r.GET(Path_Service_Name+"/match_results/audios/:audio_id", GetMatchResultByAudio)
	r.GET(Path_Service_Name+"/match_results/users/:user_id", GetMatchResultByUser)
	r.Run(":" + *httpPort)
}

func SignIn(c *gin.Context) {
	request, _ := io.ReadAll(c.Request.Body)
	info := model.SignInInfo{}
	err := json.Unmarshal(request, &info)
	if err != nil {
		fmt.Println("Error when parse json sign in info", err)
	}
	println("Try login with body ", string(request))
	info.UserName = strings.TrimSpace(info.UserName)
	if len(info.UserName) == 0 {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error UserName  is empty ", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	userInfo, _ := db.GetUserInfoByUserName(info.UserName, info.Provider)
	if userInfo.Id == 0 {
		userInfo.UserName = info.UserName
		userInfo.Provider = info.Provider
		userInfo.Avatar = info.Avatar
		userInfo.Password = info.Password
		userInfo.BirthDay = info.BirthDay
		userInfo.Gender = info.Gender
		userInfo.Name = info.Name
		userInfo, err = db.InsertUserInfo(userInfo)
		if err != nil {
			fmt.Println("Error when insert user info", err)
		}
	}
	if info.Provider == "google" && len(info.DeviceId) > 0 {
		oldInfo, _ := db.GetUserInfoByUserName(info.DeviceId, "device")
		if oldInfo.Id > 0 {
			totalOldScore := db.MergeUserInfoInRedis(oldInfo.Id, userInfo.Id)
			if totalOldScore > 0 {
				db.MergerUserInfoInMysql(oldInfo.Id, userInfo.Id)
			}
		}
	}

	expireTime := time.Now().Unix() + 7*24*60*60
	accessToken, _ := json.Marshal(model.AccessToken{User: strconv.Itoa(userInfo.Id), Expire: expireTime})
	userStat, err := db.GetUserAndStatById(userInfo.Id)
	if err != nil {
		println("error when get user stat by id ", userInfo.Id, " err ", err.Error())
	}
	totalScore, rank, err := db.GetRankByUser(userInfo.Id)
	if err != nil {
		println("error when get rank user by id ", userInfo.Id, " err ", err.Error())
	}

	responseInfo := model.SignedUserInfo{
		Id:          userInfo.Id,
		Name:        userInfo.Name,
		Avatar:      userInfo.Avatar,
		Gender:      userInfo.Gender,
		BirthDay:    userInfo.BirthDay,
		ExpireTime:  expireTime,
		AccessToken: base64.StdEncoding.EncodeToString(accessToken),
		AccuracyAvg: fmt.Sprintf("%.10f", userStat.AccuracyAvg),
		TimeAvg:     fmt.Sprintf("%.10f", userStat.TimeAvg),
		TotalScore:  fmt.Sprintf("%.10f", totalScore),
		Rank:        rank,
	}
	output, _ := json.Marshal(responseInfo)
	println("out put ", string(output))
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: responseInfo})
	c.Data(200, "text/html; charset=UTF-8", data)
}

func GetUserInfo(c *gin.Context) {
	userIdText := strings.ToLower(c.Param("user_id"))
	userId, err := strconv.Atoi(userIdText)
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when parser user id ", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}

	userInfo, err := db.GetUserById(userId)
	if err != nil {
		println("error when get user by id ", userId, " err ", err.Error())
	}
	userStat, err := db.GetUserAndStatById(userId)
	if err != nil {
		println("error when get user stat by id ", userId, " err ", err.Error())
	}
	totalScore, rank, err := db.GetRankByUser(userId)
	if err != nil {
		println("error when get rank user by id ", userId, " err ", err.Error())
	}
	responseInfo := model.BasicUserInfo{
		Id:          userInfo.Id,
		Name:        userInfo.Name,
		Avatar:      userInfo.Avatar,
		Gender:      userInfo.Gender,
		BirthDay:    userInfo.BirthDay,
		AccuracyAvg: fmt.Sprintf("%.10f", userStat.AccuracyAvg),
		TimeAvg:     fmt.Sprintf("%.10f", userStat.TimeAvg),
		TotalScore:  fmt.Sprintf("%.10f", totalScore),
		Rank:        rank,
	}
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: responseInfo})
	c.Data(200, "text/html; charset=UTF-8", data)
}

func UpdateUserInfo(c *gin.Context) {
	userIdText := strings.ToLower(c.Param("user_id"))
	userId, err := strconv.Atoi(userIdText)
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when parser user id ", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	userInfo, _ := db.GetUserById(userId)
	if userInfo.Id <= 0 {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error user id not found", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}

	forms, err := c.MultipartForm()
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "mutil part form not found", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	var avatarForm *multipart.FileHeader = nil
	if len(forms.File["avatar"]) > 0 {
		avatarForm = forms.File["avatar"][0]
	}
	name := ""
	if len(forms.Value["name"]) > 0 {
		name = forms.Value["name"][0]
		name = strings.TrimSpace(name)
	}

	avatarMd5 := ""
	if avatarForm != nil {
		avatarFile, _ := avatarForm.Open()
		avatarData, _ := io.ReadAll(avatarFile)
		if len(avatarData) > 0 {
			avatarMd5Byte := md5.Sum(avatarData)
			avatarMd5 = strings.ToLower(hex.EncodeToString(avatarMd5Byte[:]))
			avatarFileName := *avatarDir + avatarMd5 + ".png"
			println("avatar File Name ", avatarFileName, "user id ", userId)
			f, _ := os.Create(avatarFileName)
			f.Write(avatarData)
			f.Sync()
			f.Close()
		}
	}
	if len(avatarMd5) > 0 {
		newAvatar := "https://mangaverse.skymeta.pro/meme_avatars/" + avatarMd5 + ".png"
		if strings.Compare(newAvatar, userInfo.Avatar) != 0 {
			userInfo.Avatar = newAvatar
			userInfo, err = db.UpdateAvatarUserInfo(userInfo)
			if err != nil {
				println("error when update avatar user info ", err.Error())
			}
		}
	}
	if strings.Compare(name, userInfo.Name) != 0 && len(name) > 0 {
		userInfo.Name = name
		userInfo, err = db.UpdateNameUserInfo(userInfo)
		if err != nil {
			println("error when update name user info ", err.Error())
		}
	}
	responseInfo := model.BasicUserInfo{
		Id:       userInfo.Id,
		Name:     userInfo.Name,
		Avatar:   userInfo.Avatar,
		Gender:   userInfo.Gender,
		BirthDay: userInfo.BirthDay,
	}
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: responseInfo})
	c.Data(200, "text/html; charset=UTF-8", data)
}

func SignOutUserInfo(c *gin.Context) {
	userIdText := strings.ToLower(c.Param("user_id"))
	userId, err := strconv.Atoi(userIdText)
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when parser user id ", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	println("user id ", userId)
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: nil})
	c.Data(200, "text/html; charset=UTF-8", data)
}

func RemoveUserInfo(c *gin.Context) {
	userIdText := strings.ToLower(c.Param("user_id"))
	userId, err := strconv.Atoi(userIdText)
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when parser user id ", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	userInfo, _ := db.GetUserById(userId)
	if userInfo.Id <= 0 {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error user id not found", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	db.RemoveUserInfoInMysql(userId)
	db.RemoveUserInfoInRedis(userId)
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: nil})
	c.Data(200, "text/html; charset=UTF-8", data)
}

func RemoveMatchInfo(c *gin.Context) {
	userIdText := strings.ToLower(c.Param("user_id"))
	userId, err := strconv.Atoi(userIdText)
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when parser user id ", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	matchIdText := strings.ToLower(c.Param("match_id"))
	matchId, err := strconv.Atoi(matchIdText)
	if err != nil {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error when parser match id ", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	matchInfo, err := db.GetMatchResultById(matchId)
	if matchInfo.Id <= 0 {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error match id not found", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	if matchInfo.UserId != userId {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error owner not match wanted : " +strconv.Itoa(matchInfo.UserId)+	", got : "+userIdText, Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	userInfo, _ := db.GetUserById(userId)
	if userInfo.Id <= 0 {
		reposeData, _ := json.Marshal(model.HttpResponse{Code: model.HttpFail, Message: "Error user id not found", Data: nil})
		c.Data(200, "text/html; charset=UTF-8", reposeData)
		return
	}
	db.RemoveMatchInfoInMysql(matchId)
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: nil})
	c.Data(200, "text/html; charset=UTF-8", data)
}
