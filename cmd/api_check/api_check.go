package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	httpPort = flag.String("http_port", "9090", "http_port listen")
	conf     = flag.String("conf", "/home/tamnb/projects/src/github.com/nguyenbatam/crawl_pinterest_mysql/pinterest.toml", "config run file *.toml")
	c        = config.CrawlConfig{}
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
	err = db.Init(c.Postgres)
	if err != nil {
		fmt.Println("err", err)
	}
	defer db.Close()
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/topics/:topicId/images", GetImageByTopic)
	r.GET("/keywords/:keyword/images", GetImageByKeyword)
	r.Run(":" + *httpPort)
}

func GetImageByTopic(c *gin.Context) {
	topicId, err := strconv.Atoi(c.Param("topicId"))
	if err != nil {
		fmt.Println(c.Request.RequestURI)
		c.HTML(http.StatusOK, "Error path param topic not found", nil)
		return
	}
	offsetText, exit := c.GetQuery("offset")
	var offset = 0
	if exit {
		offset, err = strconv.Atoi(offsetText)
		if err != nil {
			fmt.Println(c.Request.RequestURI)
			c.HTML(http.StatusOK, "Error when offset query", offsetText)
			return
		}
	}
	lengthText, exit := c.GetQuery("length")
	var length int = 20
	if exit {
		length, err = strconv.Atoi(lengthText)
		if err != nil {
			fmt.Println(c.Request.RequestURI)
			c.HTML(http.StatusOK, "Error when length query", lengthText)
			return
		}
	}
	if length > 200 {
		length = 200
	}
	mysqlDb := db.GetDb()
	categoryName := ""

	err = mysqlDb.Model((*db.Topic)(nil)).Where("id = ?", topicId).ForEach(
		func(c *db.Topic) error {
			categoryName = c.Name
			return nil
		})
	fmt.Println(categoryName, err)

	listCrawl1 := make([]db.Image, 0)
	mysqlDb.Model((*db.ImagesInTopic)(nil)).Where("topic_id = ?", topicId).Offset(offset).Limit(length).ForEach(
		func(c *db.ImagesInTopic) error {
			image, _ := db.GetImageInfo(c.ImageId)
			listCrawl1 = append(listCrawl1, image)
			return nil
		})
	var data = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <title>Bootstrap Example</title>\n  <meta charset=\"utf-8\">\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n  <link rel=\"stylesheet\" href=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.4.1/css/bootstrap.min.css\">\n  <script src=\"https://ajax.googleapis.com/ajax/libs/jquery/3.6.0/jquery.min.js\"></script>\n  <script src=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.4.1/js/bootstrap.min.js\"></script>\n</head>\n<body>\n\n<div class=\"container\">\n " +
		" <h2>" + categoryName + "</h2>\n <table class=\"table\">\n    <tbody>\n      <tr>"

	count := 0
	for i := 0; i < len(listCrawl1); i++ {
		data = data + " <td><img src=\"" +
			listCrawl1[i].Image +
			"\" class=\"img-thumbnail\"  </td> "
		count++
		if count%5 == 0 {
			data = data + " </tr>\n <tr>"
		}
	}
	data = data + "\n</div>\n\n</body>\n</html>\n"

	c.Data(200, "text/html; charset=UTF-8", []byte(data))
}

func GetImageByKeyword(c *gin.Context) {
	keyword := c.Param("keyword")
	mysqlDb := db.GetDb()
	var crawlSource db.CrawlSource
	mysqlDb.Model((*db.CrawlSource)(nil)).Where("keyword = ?", keyword).ForEach(
		func(c *db.CrawlSource) error {
			crawlSource = *c
			return nil
		})
	listImages := make([]db.Image, 0)
	imageIds := make(map[int]bool)
	if crawlSource.Id > 0 {
		switch crawlSource.Type {
		case db.Keyword:
			mysqlDb.Model((*db.SearchKeyword)(nil)).Where("keyword_id = ?", crawlSource.Id).ForEach(
				func(c *db.SearchKeyword) error {
					if !imageIds[c.ImageId] {
						image, _ := db.GetImageInfo(c.ImageId)
						listImages = append(listImages, image)
						imageIds[c.ImageId] = true
					}
					return nil
				})
		case db.Pin:
			mysqlDb.Model((*db.RelatedPin)(nil)).Where("image_id = ?", crawlSource.ImageId).ForEach(
				func(c *db.RelatedPin) error {
					if !imageIds[c.RelatedImageId] {
						image, _ := db.GetImageInfo(c.RelatedImageId)
						listImages = append(listImages, image)
						imageIds[c.RelatedImageId] = true
					}
					return nil
				})
		}
	}
	listCrawl2 := make([]db.Image, 0)
	for i := 0; i < len(listImages); i++ {
		mysqlDb.Model((*db.RelatedPin)(nil)).Where("image_id = ?", listImages[i].Id).ForEach(
			func(c *db.RelatedPin) error {
				if !imageIds[c.RelatedImageId] {
					image, _ := db.GetImageInfo(c.RelatedImageId)
					listCrawl2 = append(listCrawl2, image)
					imageIds[c.RelatedImageId] = true
				}
				return nil
			})
	}

	var data = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <title>Bootstrap Example</title>\n  <meta charset=\"utf-8\">\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n  <link rel=\"stylesheet\" href=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.4.1/css/bootstrap.min.css\">\n  <script src=\"https://ajax.googleapis.com/ajax/libs/jquery/3.6.0/jquery.min.js\"></script>\n  <script src=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.4.1/js/bootstrap.min.js\"></script>\n</head>\n<body>\n\n<div class=\"container\">\n " +
		" <h2>" + keyword + "   depth 1 </h2>\n <table class=\"table\">\n    <tbody>\n      <tr>"

	count := 0
	for i := 0; i < len(listImages); i++ {
		data = data + " <td><img src=\"" +
			listImages[i].Image +
			"\" class=\"img-thumbnail\"  </td> "
		count++
		if count%5 == 0 {
			data = data + " </tr>\n <tr>"
		}
	}
	data = data + "</tr> </tbody> </table> "
	data = data + " <p> <h2>" + keyword + "  depth 2 </h2> </p>  \n  <table class=\"table\">\n    <tbody>\n      <tr>"
	count = 0
	for i := 0; i < len(listCrawl2); i++ {
		data = data + " <td><img src=\"" +
			listCrawl2[i].Image +
			"\" class=\"img-thumbnail\"  </td> "
		count++
		if count%5 == 0 {
			data = data + " </tr>\n <tr>"
		}
	}
	data = data + "\n</div>\n\n</body>\n</html>\n"

	c.Data(200, "text/html; charset=UTF-8", []byte(data))
}
