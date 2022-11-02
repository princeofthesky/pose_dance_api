package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	update = flag.Int("update", 0, "update info again from pinterest or not : = 1 (force) , =0 : only new info")
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
		fmt.Println("err when connect to db", err)
	}
	defer db.Close()
	mysql := db.GetDb()
	for true {
		max := time.Now().UnixNano()
		fmt.Println("Start new round update info ", time.Now())
		for true {
			imagesInfo := make([]db.Image, 0)
			mysql.Model((*db.Image)(nil)).Where("\"image\".\"crawled_time\" < ?", max).Limit(100).Order("crawled_time DESC").ForEach(
				func(c *db.Image) error {
					imagesInfo = append(imagesInfo, *c)
					return nil
				})
			if len(imagesInfo) == 0 {
				break
			}
			for _, imageInfo := range imagesInfo {
				annotation := db.Annotation{ImageId: imageInfo.Id}
				mysql.Model(&annotation).Where("\"annotation\".\"image_id\"=?", annotation.ImageId).Select()
				if *update == 1 || len(annotation.Data) == 0 {
					err := UpdatePinInfo(imageInfo)
					if err != nil {
						fmt.Println("err when update pin info", imageInfo.Id, imageInfo.SourceId)
						fmt.Println("err", err)
					}
				}
				max = imageInfo.CrawledTime
			}
		}
		time.Sleep(20 * time.Minute)
	}
}

func UpdatePinInfo(imagesInfo db.Image) error {
	resp, err := http.Get("https://www.pinterest.com/pin/" + imagesInfo.SourceId)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to fetch data: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	data := doc.Find("#__PWS_DATA__").Text()
	fmt.Println(imagesInfo.Id, imagesInfo.SourceId)
	//ioutil.WriteFile("1.txt", []byte(data), fs.ModePerm)
	response := map[string]interface{}{}
	err = json.Unmarshal([]byte(data), &response)
	if err != nil {
		return err
	}
	props := response["props"].(map[string]interface{})
	initialReduxState := props["initialReduxState"].(map[string]interface{})
	pins := initialReduxState["pins"].(map[string]interface{})
	if pins[imagesInfo.SourceId] == nil {
		return fmt.Errorf("can request data with Pin ID :%v", imagesInfo.SourceId)
	}

	info := pins[imagesInfo.SourceId].(map[string]interface{})
	title := ""
	if info["title"] != nil {
		title = info["title"].(string)
	}
	if len(title) == 0 {
		if info["grid_title"] != nil {
			grid_title := info["grid_title"].(string)
			if len(grid_title) > 0 {
				title = grid_title
			}
		}

	}
	description := ""
	if info["description"] != nil {
		description = info["description"].(string)
	}
	closeupUnifiedDescription := info["closeup_unified_description"].(string)
	if len(title) == 0 {
		if len(description) > 0 {
			title = description
		} else if len(closeupUnifiedDescription) > 0 {
			title = closeupUnifiedDescription
		} else {
			title = ""
		}
	}
	if len(description) == 0 {
		if len(info["description_html"].(string)) > 0 {
			description = info["description_html"].(string)
		} else {
			description = ""
		}
	}
	imagesInfo.Title = title
	imagesInfo.Description = description
	var closeupAttribution map[string]interface{}
	if info["closeup_attribution"] != nil {
		closeupAttribution = info["closeup_attribution"].(map[string]interface{})
	} else {
		closeupAttribution = info["pinner"].(map[string]interface{})
	}

	imagesInfo.OwnerName = closeupAttribution["full_name"].(string)
	imagesInfo.OwnerUrl = "https://www.pinterest.com/" + closeupAttribution["username"].(string)

	annotations := make([]string, 0)
	if info["pin_join"] != nil {
		pinJoin := info["pin_join"].(map[string]interface{})
		if pinJoin["annotations_with_links"] != nil {
			annotationsWithLinks := pinJoin["annotations_with_links"].(map[string]interface{})
			for key, _ := range annotationsWithLinks {
				annotations = append(annotations, key)
			}
		}
	}
	//hashtags := make([]string, 0)
	//if info["hashtags"] != nil {
	//	hashtagsInfo := info["hashtags"].([]interface{})
	//	for _, key := range hashtagsInfo {
	//		hashtags = append(hashtags, key.(string))
	//	}
	//}
	boards := initialReduxState["boards"].(map[string]interface{})
	imagesInfo.BoardDescription = ""
	for _, v := range boards {
		board := v.(map[string]interface{})
		if board["description"] != nil {
			imagesInfo.BoardDescription = board["description"].(string)
		}
		break
	}
	err = db.UpdateImageInfo(imagesInfo)
	if err != nil {
		fmt.Println("error when update image info ", imagesInfo.Id, err)
	}
	annotationJson, _ := json.Marshal(annotations)
	db.InsertAnnotationImage(db.Annotation{imagesInfo.Id, string(annotationJson)})
	return nil
}
