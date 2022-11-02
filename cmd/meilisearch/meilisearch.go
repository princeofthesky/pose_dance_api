package main

import (
	"flag"
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"go-pinterest/db"
	"strconv"
	"time"
)

var (
	database = "meme-images"
	url      = flag.String("url", "http://127.0.0.1:7700", "url query meili meilisearch ")
)

func main() {
	db.Init()
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: *url,
	})
	dbSearch := client.Index(database)
	ok, _ := dbSearch.DeleteAllDocuments()
	for true {
		check,err:=dbSearch.GetTask(ok.TaskUID)
		fmt.Println("delete all document with result ", check.Status, err)
		if check.Status == meilisearch.TaskStatusSucceeded {
			break
		}
		time.Sleep(5 * time.Second)
	}
	categories, _ := db.GetAllCategory()
	for _, category := range categories {
		keywords, _ := db.GetAllKeywordByCategory(category)
		for _, keyword := range keywords {
			for depth := 1; depth < 3; depth++ {
				imageIds, _ := db.GetAllImageIdByCategoryAndDepth(category, keyword, depth)
				for _, imageId := range imageIds {
					id, _ := strconv.ParseInt(imageId, 10, 64)
					imageInfo, _ := db.GetImageInfo(id)
					dataSearch := make(map[string]interface{})
					dataSearch["id"] = imageInfo.Id
					dataSearch["poster"] = imageInfo.Image
					dataSearch["title"] = imageInfo.Title
					dataSearch["keywords"] = imageInfo.KeyWords
					dataSearch["description"] = imageInfo.Description
					dataSearch["hashtags"] = imageInfo.Hashtags
					dataSearch["annotations"] = imageInfo.Annotations
					dataSearch["board_description"] = imageInfo.BoardDescription
					dataSearch["depth"] = depth
					_, err := dbSearch.AddDocuments(dataSearch,"id")
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}
