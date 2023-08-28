package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/date"
	"go-pinterest/db"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAvatar     = "default.png"
	Path_Service_Name = "/pose_dance_admin/"
	DateLayout        = "2006-01-02"
)

var (
	httpPort          = flag.String("http_port", "9082", "http_port listen")
	conf              = flag.String("conf", "./pose_dance_api.toml", "config run file *.toml")
	c                 = config.CrawlConfig{}
	videoDir          = flag.String("video_dir", "/meme_videos/", "video meme direction")
	videoCoverDir     = flag.String("cover_dir", "/meme_covers/", "cover meme direction")
	avatarDir         = flag.String("avatar_dir", "/meme_avatars/", "avatar meme direction")
	version           = flag.Int("version", 0, "= 0 , server , 1 local")
	loc, _            = time.LoadLocation("Asia/Ho_Chi_Minh")
	baseUrlLocal      = "http://10.0.1.130/"
	allSubVideoFolder = map[string]bool{}
	allSubCoverFolder = map[string]bool{}
)

func main() {
	flag.Parse()
	baseUrlLocal = baseUrlLocal
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
	go func() {
		for true {
			InitListAudio()
			time.Sleep(5 * time.Minute)
		}
	}()
	r := gin.Default()
	InitListAudio()
	LoadAllSubVideoFolder()
	LoadAllSubCoverFolder()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	// leader board
	r.GET(Path_Service_Name+"/videos", GetListVideoMatchResultByDay)
	r.GET(Path_Service_Name+"/videos/filters", GetListVideoMatchResultByAudioAndScore)
	r.GET(Path_Service_Name+"/videos/filters/json", GetDataJsonListVideoMatchResultByAudioAndScore)
	r.GET(Path_Service_Name+"/videos/users", GetListVideoByVideoUser)
	r.GET(Path_Service_Name+"/users/:user_id/videos", GetListVideoByUser)
	r.GET(Path_Service_Name+"/users/top", GetTopUser)
	r.GET(Path_Service_Name+"/audios/:audio_id/videos", GetListVideoByAudio)

	r.Run(":" + *httpPort)

}

func LoadAllSubVideoFolder() {
	filepath.Walk(*videoDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			allSubVideoFolder[path] = true
			return nil
		}
		return nil
	})
}

func LoadAllSubCoverFolder() {
	filepath.Walk(*videoCoverDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			allSubCoverFolder[path] = true
			return nil
		}
		return nil
	})
}
func FileIsExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	return true
}

func UpdateVideoLink(match db.MatchResult) db.MatchResult {
	videoFile := *videoDir + match.VideoMd5 + ".mp4"
	if FileIsExist(videoFile) {
		match.Video = baseUrlLocal + strings.ReplaceAll(videoFile, *videoDir, "meme_videos/")
		return match
	}
	folderDaily := date.GetFolderDaily(match.ReceivedTime)
	videoFile = *videoDir + "/" + folderDaily + "/" + match.VideoMd5 + ".mp4"
	if FileIsExist(videoFile) {
		match.Video = baseUrlLocal + strings.ReplaceAll(videoFile, *videoDir, "meme_videos/")
		return match
	}
	for folder, _ := range allSubVideoFolder {
		videoFile = folder + "/" + match.VideoMd5 + ".mp4"
		if FileIsExist(videoFile) {
			match.Video = baseUrlLocal + strings.ReplaceAll(videoFile, *videoDir, "meme_videos/")
			return match
		}
	}
	return match
}

func UpdateCoverLink(match db.MatchResult) db.MatchResult {
	match.Cover = strings.TrimSpace(match.Cover)

	coverFile := *videoCoverDir + match.CoverMd5 + ".png"
	if FileIsExist(coverFile) {
		match.Cover = baseUrlLocal + strings.ReplaceAll(coverFile, *videoCoverDir, "meme_covers/")
		return match
	}
	folderDaily := date.GetFolderDaily(match.ReceivedTime)
	coverFile = *videoCoverDir + "/" + folderDaily + "/" + match.CoverMd5 + ".png"
	if FileIsExist(coverFile) {
		match.Cover = baseUrlLocal + strings.ReplaceAll(coverFile, *videoCoverDir, "meme_covers/")
		return match
	}
	for folder, _ := range allSubCoverFolder {
		coverFile = folder + "/" + match.CoverMd5 + ".png"
		if FileIsExist(coverFile) {
			match.Cover = baseUrlLocal + strings.ReplaceAll(coverFile, *videoCoverDir, "meme_covers/")
			return match
		}
	}
	return match
}

func UpdateLink(matches []db.MatchResult) []db.MatchResult {
	if *version == 0 {
		return matches
	}
	updatedMatches := []db.MatchResult{}
	for _, match := range matches {
		match = UpdateVideoLink(match)
		match = UpdateCoverLink(match)
		updatedMatches = append(updatedMatches, match)
	}
	return updatedMatches
}
func GetDocumenHtml(title string, matchs []db.MatchResult, nextPage string) string {
	matchs = UpdateLink(matchs)
	html := "<!DOCTYPE html> <html><style> table, th, td { border:1px solid black;}</style><body><h2>" +
		title +
		"</h2> " +
		"<table align=\"center\" style=\"width:90%; max-width:1000px;\"> <thead> <tr> <th>Title</th> <th>Score</th> <th>Author</th> <th>Time</th><th>Video</th> </tr> </thead> <tbody>  "

	for i := 0; i < len(matchs); i++ {
		match := matchs[i]
		user, err := db.GetUserById(match.UserId)
		if err != nil {
			println("error when get user by id ", match.UserId, err)
			continue
		}

		html = html + "<tr>  <td> <a href=\"" + "/pose_dance_admin/audios/" +
			strconv.Itoa(match.AudioId) +
			"/videos \"> " +
			MapSongs[strconv.Itoa(match.AudioId)].Info.AudioTitle +
			"</a> " +
			"</td> <td>" +
			strconv.Itoa(match.Score) +
			"</td>   <td>  <a href=\"" + "/pose_dance_admin/users/" +
			strconv.Itoa(match.UserId) +
			"/videos \"> " +
			user.Name + "</td> <td> " +
			time.Unix(match.ReceivedTime, 0).In(loc).Format(time.RFC3339) +
			"</td> <td> " +
			"<video width=\"320\" height=\"240\" controls  preload=\"none\" poster=\"" +
			match.Cover +
			"\">  <source src=\"" +
			match.Video + "\"  autostart=\"false\"  type=\"video/mp4\"> </video>" + "</td> </tr> "
	}
	html = html + "</tbody></table> \n <a href=\"" + nextPage + "\"> <button class=\"btn btn-primary btn-lg\">Next Page</button>  </a> </body> </html> "
	return html
}

func GetListVideoMatchResultByDay(c *gin.Context) {
	dayText, _ := c.GetQuery("day")
	statDate, err := time.Parse(DateLayout, dayText)
	if err != nil {
		println("err when pare date :", dayText, "err", err.Error())
		today := time.Now().Format(DateLayout)
		statDate, _ = time.Parse(DateLayout, today)
	}
	offsetText, _ := c.GetQuery("offset")
	offset, _ := strconv.ParseInt(offsetText, 10, 64)
	if offset <= 0 {
		offset = statDate.Add(24 * time.Hour).Unix()
	}

	matchs, err := db.GetTopVideoMatchResultByTime(offset)
	if err != nil {
		println("error when get list new video ", err.Error())
	}

	title := "Day : " + statDate.Format(DateLayout) + " , Last Time : " + time.Unix(offset, 0).Format(time.RFC3339)
	nextPage := ""
	if len(matchs) >= 20 {
		nextPage = "/pose_dance_admin/videos?day=" + statDate.Format(DateLayout) + "&offset=" + strconv.FormatInt(matchs[len(matchs)-1].ReceivedTime, 10)
	}
	html := GetDocumenHtml(title, matchs, nextPage)
	c.Data(200, "text/html; charset=UTF-8", []byte(html))
}

func GetListVideoMatchResultByAudioAndScore(c *gin.Context) {
	lowerScoreText, _ := c.GetQuery("lower_score")
	lowerScore, err := strconv.Atoi(lowerScoreText)
	if err != nil {
		lowerScore = 0
	}
	audioIdText, _ := c.GetQuery("audio_id")
	audioId, err := strconv.Atoi(audioIdText)
	if err != nil {
		audioId = 0
	}
	sortType, _ := c.GetQuery("type")
	if sortType != "score" {
		sortType = "received_time"
	}
	offsetText, _ := c.GetQuery("offset")
	offset, _ := strconv.Atoi(offsetText)
	if offset <= 0 {
		offset = 0
	}

	matchs, err := db.GetTopVideoMatchResultByAudioIdAndLimitScore(audioId, "",lowerScore, sortType, offset)
	if err != nil {
		println("error when get list new video ", err.Error())
	}

	title := "AudioId  : " + audioIdText + " , Off_set: " + offsetText
	nextPage := ""
	if len(matchs) >= 20 {
		nextPage = "/pose_dance_admin/videos/filters?lower_score=" + lowerScoreText + "&audio_id=" + audioIdText + "&type=" + sortType + "&offset=" + strconv.Itoa(offset+20)
	}
	html := GetDocumenHtml(title, matchs, nextPage)
	c.Data(200, "text/html; charset=UTF-8", []byte(html))
}

func GetDataJsonListVideoMatchResultByAudioAndScore(c *gin.Context) {
	lowerScoreText, _ := c.GetQuery("lower_score")
	lowerScore, err := strconv.Atoi(lowerScoreText)
	if err != nil {
		lowerScore = 0
	}
	audioIdText, _ := c.GetQuery("audio_id")
	audioId, err := strconv.Atoi(audioIdText)
	if err != nil {
		audioId = 0
	}
	sortType, _ := c.GetQuery("type")
	if sortType != "score" {
		sortType = "received_time"
	}
	limitText, _ := c.GetQuery("limit")
	limit, _ := strconv.Atoi(limitText)
	if limit <= 0 {
		limit = 20
	}
	offsetText, _ := c.GetQuery("offset")
	offset, _ := strconv.Atoi(offsetText)
	if offset <= 0 {
		offset = 0
	}

	startTimeText, _ := c.GetQuery("start_time")
	startTime, _ := strconv.Atoi(startTimeText)
	if startTime <= 0 {
		startTime = 0
	}

	endTimeText, _ := c.GetQuery("end_time")
	endTime, _ := strconv.ParseInt(endTimeText, 10, 64)
	if endTime <= 0 {
		endTime = time.Now().Unix()
	}

	matchs, err := db.GetTopVideoMatchResultByAudioIdAndLimitScoreInATime(audioId, lowerScore, sortType, startTime, endTime, offset, limit)
	if err != nil {
		println("error when get list new video ", err.Error())
	}
	matchs = UpdateLink(matchs)
	response, _ := json.Marshal(matchs)
	c.Data(200, "text/html; charset=UTF-8", []byte(response))
}

func GetListVideoByVideoUser(c *gin.Context) {
	matchIdText, _ := c.GetQuery("video_id")
	matchId, _ := strconv.Atoi(matchIdText)
	file, _ := c.GetQuery("file")
	offsetText, _ := c.GetQuery("offset")
	offset, _ := strconv.ParseInt(offsetText, 10, 64)
	if offset <= 0 {
		offset = time.Now().Unix()
	}
	userId, err := db.GetUserIdByVideoUser(file)
	if matchId > 0 {
		matchResult, _ := db.GetMatchResultById(matchId)
		if matchResult.Id > 0 {
			userId = matchResult.UserId
		}
	}
	matches, err := db.GetNewVideoMatchResultByUser(userId, offset, 20)
	if err != nil {
		println("error when get list new video ", err.Error())
	}
	user, err := db.GetUserById(userId)
	if err != nil {
		println("error when get user id ", err.Error())
	}
	title := " User ID : " + strconv.Itoa(user.Id) + " ;  name : " + user.Name
	nextPage := ""
	if len(matches) >= 20 {
		nextPage = "/pose_dance_admin/users/" + strconv.Itoa(user.Id) + "/videos" + "?offset=" + strconv.FormatInt(matches[len(matches)-1].ReceivedTime, 10)
	}
	html := GetDocumenHtml(title, matches, nextPage)
	c.Data(200, "text/html; charset=UTF-8", []byte(html))
}

func GetListVideoByUser(c *gin.Context) {
	userIdText := c.Param("user_id")
	userId, _ := strconv.Atoi(userIdText)
	offsetText, _ := c.GetQuery("offset")
	offset, _ := strconv.ParseInt(offsetText, 10, 64)
	if offset <= 0 {
		offset = time.Now().Unix()
	}
	matchs, err := db.GetNewVideoMatchResultByUser(userId, offset, 20)
	if err != nil {
		println("error when get list new video ", err.Error())
	}
	user, err := db.GetUserById(userId)
	if err != nil {
		println("error when get user id ", err.Error())
	}
	title := " User ID : " + strconv.Itoa(user.Id) + " ;  name : " + user.Name
	nextPage := ""
	if len(matchs) >= 20 {
		nextPage = "/pose_dance_admin/users/" + userIdText + "/videos" + "?offset=" + strconv.FormatInt(matchs[len(matchs)-1].ReceivedTime, 10)
	}
	html := GetDocumenHtml(title, matchs, nextPage)
	c.Data(200, "text/html; charset=UTF-8", []byte(html))
}

func GetTopUser(c *gin.Context) {
	limitUserText, _ := c.GetQuery("limit_user")
	limitUser, _ := strconv.Atoi(limitUserText)
	if limitUser <= 0 {
		limitUser = 10
	}

	limitMatchText, _ := c.GetQuery("limit_match")
	limitMatch, _ := strconv.Atoi(limitMatchText)
	if limitMatch <= 0 {
		limitMatch = 10
	}

	startTimeText, _ := c.GetQuery("start_time")
	startTime, _ := strconv.ParseInt(startTimeText, 10, 64)
	if startTime <= 0 {
		startTime = 0
	}
	endTimeText, _ := c.GetQuery("end_time")
	endTime, _ := strconv.ParseInt(endTimeText, 10, 64)
	if endTime <= 0 {
		endTime = time.Now().Unix()
	}
	output, _ := c.GetQuery("output")
	scores, bestMatchs := db.GetTopUserByMaxScore(limitUser, limitMatch, startTime, endTime)
	if output == "json" {
		type BestUser struct {
			TotalScore int            `json:"total_score"`
			BestMatch  db.MatchResult `json:"best_match"`
		}
		bestMatchs = UpdateLink(bestMatchs)
		bestUsers := []BestUser{}
		for index, match := range bestMatchs {
			bestUsers = append(bestUsers, BestUser{TotalScore: scores[index], BestMatch: match})
		}
		response, _ := json.Marshal(bestUsers)
		c.Data(200, "text/html; charset=UTF-8", []byte(response))
		return
	}

	matchs := []db.MatchResult{}
	for index, _ := range bestMatchs {
		matchs = append(matchs, db.MatchResult{
			Score:        scores[index],
			AudioId:      bestMatchs[index].AudioId,
			UserId:       bestMatchs[index].UserId,
			Video:        bestMatchs[index].Video,
			Cover:        bestMatchs[index].Cover,
			ReceivedTime: bestMatchs[index].ReceivedTime,
		})
	}
	title := " Top User  limit " + limitUserText + " ;  limit match  : " + limitMatchText + " ; start time : " + startTimeText + " ; end time text : " + endTimeText
	nextPage := ""
	//if len(matchs) >= 20 {
	//	nextPage = "/pose_dance_admin/users/" + userIdText + "/videos" + "?offset=" + strconv.FormatInt(matchs[len(matchs)-1].ReceivedTime, 10)
	//}
	html := GetDocumenHtml(title, matchs, nextPage)
	c.Data(200, "text/html; charset=UTF-8", []byte(html))
}

func GetListVideoByAudio(c *gin.Context) {
	audioId := c.Param("audio_id")
	offsetText, _ := c.GetQuery("offset")
	offset, _ := strconv.ParseInt(offsetText, 10, 64)
	if offset <= 0 {
		offset = time.Now().Unix()
	}
	matchs, err := db.GetNewVideoMatchResultByAudioId(audioId, offset, 20)
	if err != nil {
		println("error when get list new video ", err.Error())
	}
	if err != nil {
		println("error when get user id ", err.Error())
	}
	song := MapSongs[audioId]
	title := " audio ID : " + audioId + " ;  title : " + song.Info.AudioTitle
	nextPage := ""
	if len(matchs) >= 20 {
		nextPage = "/pose_dance_admin/audios/" + audioId + "/videos" + "?offset=" + strconv.FormatInt(matchs[len(matchs)-1].ReceivedTime, 10)
	}
	html := GetDocumenHtml(title, matchs, nextPage)
	c.Data(200, "text/html; charset=UTF-8", []byte(html))
}
