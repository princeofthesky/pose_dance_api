package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var (
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
	endTime := time.Now().Unix()
	startTime := endTime - 24*60*60

	mysqlDb := db.GetMysqlDb()

	mysqlDb.Model((*db.MatchResult)(nil)).Column("*").Where("received_time > ? AND received_time < ? AND length(cover)>0 ", startTime, endTime).ForEach(
		func(c *db.MatchResult) error {
			file := strings.ReplaceAll(c.Cover, "https://mangaverse.skymeta.pro", "")
			splitTexts := strings.Split(file, "/")
			folder := ""
			coverFile := splitTexts[len(splitTexts)-1]
			for i := 1; i < len(splitTexts)-1; i++ {
				folder = folder + "/" + splitTexts[i]
			}
			_, err := os.Stat(file)
			if err == nil {
				return nil
			}
			os.MkdirAll(folder, os.ModePerm)
			cmdLine := "cp /meme_covers/" + coverFile + "  " + file
			println(cmdLine)

			return nil
		})

}
