package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"io/fs"
	"io/ioutil"
	"net/http"
)

var (
	source = flag.String("source", "/home/tamnb/projects/src/github.com/nguyenbatam/crawl_pinterest_mysql/sample_crawl_source.csv", "source category Pinterest")
	conf   = flag.String("conf", "/home/tamnb/projects/src/github.com/nguyenbatam/crawl_pinterest_mysql/pinterest.toml", "config run file *.toml")
	c      = config.CrawlConfig{}
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
	mysqlDb := db.GetDb()

	mysqlDb.Model((*db.Image)(nil)).Column("*").Where("image LIKE '%gif%'").ForEach(
		func(c *db.Image) error {
			imageGifInfo, err := GetGifInfo(c.Id, c.SourceId)
			if err != nil {
				fmt.Println("error when get gif info", err)
				return err
			}
			_, err = mysqlDb.Model(&imageGifInfo).Insert()
			fmt.Println(imageGifInfo, err)
			return nil
		})

}

func GetGifInfo(imageId int, pinID string) (db.ImageGif, error) {
	imageGif := db.ImageGif{ImageId: imageId}
	resp, err := http.Get("https://www.pinterest.com/pin/" + pinID)
	if err != nil {
		return imageGif, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return imageGif, fmt.Errorf("failed to fetch data: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return imageGif, err
	}
	data := doc.Find("#__PWS_DATA__").Text()

	ioutil.WriteFile("1.txt", []byte(data), fs.ModePerm)
	response := map[string]interface{}{}
	err = json.Unmarshal([]byte(data), &response)
	if err != nil {
		return imageGif, err
	}
	props := response["props"].(map[string]interface{})
	initialReduxState := props["initialReduxState"].(map[string]interface{})
	pins := initialReduxState["pins"].(map[string]interface{})
	if pins[pinID] == nil {
		return imageGif, fmt.Errorf("can request data with Pin ID :%v", pinID)
	}
	info := pins[pinID].(map[string]interface{})
	images := info["images"].(map[string]interface{})

	for _typeImage, _ := range images {
		sizeValue := images[_typeImage].(map[string]interface{})
		if _typeImage == "orig" {
			imageGif.Url = sizeValue["url"].(string)
			imageGif.Width = int(sizeValue["width"].(float64))
			imageGif.Height = int(sizeValue["height"].(float64))
		}
	}
	return imageGif, nil
}
