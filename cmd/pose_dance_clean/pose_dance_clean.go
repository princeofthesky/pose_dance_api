package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"go-pinterest/model"
	"io"
	"io/fs"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	conf             = flag.String("conf", "./pose_dance_api.toml", "config run file *.toml")
	c                = config.CrawlConfig{}
	videoDir         = flag.String("video_dir", "/meme_videos/", "video meme direction")
	videoCoverDir    = flag.String("cover_dir", "/meme_covers/", "cover meme direction")
	avatarDir        = flag.String("avatar_dir", "/meme_avatars/", "avatar meme direction")
	maxDay           = flag.Int("max_day", 14, "maximum day ")
	lowerScore       = flag.Int("lower_score", 10, "remove if score <= lower")
	maxVideoPerAudio = flag.Int("max_video", 1000, "max video per audio")
	DateLayout       = "2006-01-02"
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
		fmt.Println("err when connect postgres", err.Error())
	}
	defer db.Close()
	InitListAudio()
	TopVideos := map[string]int{}
	TopCovers := map[string]int{}
	lowestScoresInTop := map[string]int{}
	limitTime := int(time.Now().Unix() - int64(*maxDay)*24*60*60)
	for audioId, _ := range MapSongs {
		//audioIdInt, _ := strconv.Atoi(audioId)
		//totalMatch, err := db.GetTotalMatchResultByAudioId(audioId)
		if err != nil {
			fmt.Println("err when get total match by audio Id ", audioId, err.Error())
		}
		matchResults := []db.MatchResult{}
		err = db.GetMysqlDb().Model((*db.MatchResult)(nil)).Column("*").
			Where("audio_id = ?  AND length(video)>0 AND received_time >= ? ", audioId, limitTime).
			Order("received_time DESC").ForEach(
			func(c *db.MatchResult) error {
				matchResults = append(matchResults, *c)
				return nil
			})
		if err != nil {
			fmt.Println("err when get newest video by audio Id ", audioId, err.Error())
		}
		for _, result := range matchResults {
			fileNames := strings.Split(result.Video, "/")
			TopVideos[fileNames[len(fileNames)-1]] = result.Id
			fileNames = strings.Split(result.Cover, "/")
			TopCovers[fileNames[len(fileNames)-1]] = result.Id
		}
		println("audio ", audioId,
			"newest ", *maxDay, " day ==> ", len(matchResults))
		matchResults, err = db.GetBestMatchResultByAudioId(audioId, *maxVideoPerAudio)
		if err != nil {
			fmt.Println("err when get top match by audio Id ", audioId, err.Error())
		}
		minScore := math.MaxInt16
		for _, result := range matchResults {
			fileNames := strings.Split(result.Video, "/")
			TopVideos[fileNames[len(fileNames)-1]] = result.Id
			fileNames = strings.Split(result.Cover, "/")
			TopCovers[fileNames[len(fileNames)-1]] = result.Id
			if result.Score < minScore {
				minScore = result.Score
			}
		}
		lowestScoresInTop[audioId] = minScore
		println("audio ", audioId, " top ", *maxVideoPerAudio, "minScore", minScore, " ==> ", len(matchResults))
	}
	println("total safe", len(TopVideos))
	totalVideo := 0
	totalSize := int64(0)
	os.Remove("clean_video_cover.sh")
	os.Remove("clean_video_cover_log.txt")
	removeCmd, err := os.Create("clean_video_cover.sh")
	removeLog, err := os.Create("clean_video_cover_log.txt")
	defer removeCmd.Close()
	defer removeLog.Close()
	filepath.Walk("/meme_videos/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		dateFolderText := strings.Split(path, "/")
		if len(dateFolderText) > 4 {
			dateFolder := dateFolderText[2] + "-" + dateFolderText[3] + "-" + dateFolderText[4]
			dateFolderDay, err := time.Parse(DateLayout, dateFolder)
			if err != nil {
				removeLog.WriteString("err when parse date folder " + err.Error() + "\n")
				return nil
			}
			if int(dateFolderDay.Unix()) > limitTime {
				removeLog.WriteString("not remove file " + path + "\n")
				return nil
			}
		}
		fileName := info.Name()
		if TopVideos[fileName] > 0 || TopCovers[fileName] > 0 {
			return nil
		}
		match, err := db.GetMatchResultByVideoFileOrCoverFile(fileName)
		if err != nil {
			removeLog.WriteString("err when get match result by file  " + fileName + "\t" + err.Error() + "\n")
			return nil
		}
		if match.Id == 0 {
			removeLog.WriteString("not found match result by file  " + fileName + "\n")
			return nil
		}

		match.Video = ""
		_, err = db.UpdateVideoMatchResult(match)
		if err != nil {
			removeLog.WriteString("err when update video match result " + strconv.Itoa(match.Id) + "\t" + err.Error() + "\n")
		}
		removeLog.WriteString(strconv.Itoa(match.Id) + "\t " + fileName +
			" \t" + strconv.Itoa(match.Score) + "\t" + strconv.Itoa(lowestScoresInTop[strconv.Itoa(match.AudioId)]) + "\t" +
			strconv.FormatBool(match.Score < lowestScoresInTop[strconv.Itoa(match.AudioId)]) + " \t " +
			" \t " + time.Unix(match.ReceivedTime, 0).String() + "\n")
		removeCmd.WriteString("rm " + path + "\n")
		totalSize = totalSize + info.Size()
		//fmt.Println("rm "+path, info.ModTime().String())
		totalVideo++
		return nil
	})

	totalCover := 0
	filepath.Walk("/meme_covers/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		dateFolderText := strings.Split(path, "/")
		if len(dateFolderText) > 4 {
			dateFolder := dateFolderText[2] + "-" + dateFolderText[3] + "-" + dateFolderText[4]
			dateFolderDay, err := time.Parse(DateLayout, dateFolder)
			if err != nil {
				removeLog.WriteString("err when parse date folder " + err.Error() + "\n")
				return nil
			}
			if int(dateFolderDay.Unix()) > limitTime {
				removeLog.WriteString("not remove file " + path + "\n")
				return nil
			}
		}

		fileName := info.Name()
		if TopVideos[fileName] > 0 || TopCovers[fileName] > 0 {
			return nil
		}
		match, err := db.GetMatchResultByVideoFileOrCoverFile(fileName)
		if err != nil {
			fmt.Println("err when get match result by file  ", fileName, err.Error())
			return nil
		}
		if match.Id == 0 {
			removeLog.WriteString("not found match result by file  " + fileName + "\n")
			return nil
		}
		match.Cover = ""
		_, err = db.UpdateVideoMatchResult(match)
		if err != nil {
			fmt.Println("err when update cover match result ", match.Id, err.Error())
		}
		removeLog.WriteString(strconv.Itoa(match.Id) + "\t " +
			fileName +
			" \t" + strconv.Itoa(match.Score) + "\t" + strconv.Itoa(lowestScoresInTop[strconv.Itoa(match.AudioId)]) + "\t" +
			strconv.FormatBool(match.Score < lowestScoresInTop[strconv.Itoa(match.AudioId)]) + " \t " +
			time.Unix(match.ReceivedTime, 0).String() + "\n")
		removeCmd.WriteString("rm " + path + "\n")
		totalSize = totalSize + info.Size()
		totalCover++
		return nil
	})

	println("totalVideo remove", totalVideo)
	println("totalCover remove", totalCover)
	println("totalSize remove", totalSize)
}
