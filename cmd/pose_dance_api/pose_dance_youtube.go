package main

import (
	"fmt"
	"go-pinterest/db"
	"go-pinterest/model"
)

var MapYoutubeVideos = map[int]*model.YoutubeVideoInfo{}
var ListYoutubeMatchResult []db.MatchResult

func InitListYoutubeVideo() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error when reload list youtube video ", r)
		}
	}()
	allYoutubeMatch, err := db.GetAllYoutubeMatch()
	if err != nil {
		println("error when get all youtube match", err.Error())
	}
	var newMapYoutubes = map[int]*model.YoutubeVideoInfo{}
	var newListYoutubeMatchResult = []db.MatchResult{}
	for _, youtubeMatch := range allYoutubeMatch {
		videoInfo := model.YoutubeVideoInfo{
			MatchId:   youtubeMatch.MatchId,
			Id:        youtubeMatch.YoutubeId,
			ShortUrl:  "https://www.youtube.com/shorts/" + youtubeMatch.YoutubeId,
			Thumbnail: youtubeMatch.Thumbnail,
			Url:       "https://www.youtube.com/watch?v=" + youtubeMatch.YoutubeId,
		}
		match, err := db.GetMatchResultById(youtubeMatch.MatchId)
		if err != nil {
			continue
		}
		newMapYoutubes[youtubeMatch.MatchId] = &videoInfo
		newListYoutubeMatchResult = append(newListYoutubeMatchResult, match)
	}
	if len(newMapYoutubes) > 0 {
		MapYoutubeVideos = newMapYoutubes
		ListYoutubeMatchResult = newListYoutubeMatchResult
	}
}
