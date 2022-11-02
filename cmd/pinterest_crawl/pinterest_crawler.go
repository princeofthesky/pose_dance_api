package main

import (
	"bytes"
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
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	//SessionKey    = "TWc9PSZ5azdObjM3eEJJTVlJQUNjTStlSzU3ajhGWC8veUI3T1VycUVrc3Y4dmkzVlgvei9ZTW91T1JiTzJycTZHMHRJVjBDZ1hCSWMwSERSREpFUkdvNy9wQStZRGF2VktkVVhKN2V1LzU1WGpKa0hpYzFidG96UkdZQXdITWFDenZJVVZ2dVl3aVBHT1IyK3NiZnJJR0hWb1plRER5Z2hPM1RHc1BhWWVhdUlLbjg1MlkzTnVDTTVZNjFtbjBYQndQVmF0THhWVk85b29hMFV2OUx6NVRFMkFpZlAyZStaU0dCQUovQWVtQTVuR2xDNm9NaU0yL3RTdXV5aGhXVXprN0dOOU1EU3lVcy9VMUk2SVpzK0FHYUE5TExLZGFUVmY5aEdJQVJmcjJVUnBlc3ZhVENxRzFoY3A1UHFOSDlIYXdZM1RpQ2sxb2FKdXJTZ2pGMlRvNzdUZFFOK2ZlTXZVdGx0cnFFUWpuSzVtTU1GTzhWejZKcWw3all0VlM1aVRQcFdZdHpQK0Q4MGdGZUhEMnN0d0xURThOY2gwMUt6QXpDRVZJcllSQXBVZmhKMzROUTUyU21EclE1VE5Lb1E3TVlPSU9zZE5xRlNIcnJ3eXhWc2xIbTBvcnBnSHlLMlBYQnEwYm5kQSt4dXVMK0VTTmVqMWpGQThzUjdiTDdNU1lXVFhWaVpxQ3hEK3E0R2ZWZlo3ek1selBWbDJiN3VnMlJrRE1DTXRYeXJDeGhWK0dqeGJqRWVWNTdBNW9ZcWkvdmRQU0NNMmszZlNKaGtBSG1icFExdkF6Mjh4TDFydXFyNk5GZmZTSTkyaVhYaE1XeTZ6a1JDaEFzdlViTWVyajRpcWo5Z2xKOTBBOU9HNEdxVGw2MGR5STlFdGVXYnRtaVJ6ZTY2OGdqNVNiMkhpUHZrL0VldnlnMlVhV0pKQkZnenE1Q3RXMkI2Qi9WTVUxK2VTdUoyam1hWVB2Um9Obyt0S2cvQUxrM3Y4dWxaMHJmMmdLVlJMLytHUWFYNVoxWElqSk1kRnlnOEpHQ0JscVZWN0FjNHJ1aDBTK1JlenRUODdVSUx1RmtZOERlUEFaYllETld4eit0ZU8vL3RGcFpST0lSYWxKcHJTSFI5WmRab1hRR2JHM0c4dHNmdTlmVFk0Y2FVeDkwbm9xWk1KSTNDb28rWTMxSjVaZnZCN0kxdWIzWGEwVCswZDVwV3BIVnh4eEpyRUdRVEovT3dTRW5WdW05U20vd1hTTCtld0piRERRZGN0ckRoa0NKS212K1gmeHlYTkcvdFJNZmtBNzJDTUQ1SUt3cmJFUFlFPQ=="
	//Csrftoken     = "1d7df6fbd5e80a9b140abd4196ccf0d6"
	//maxRelated    = flag.Int("max_related_pin", 50, "max items per query related")
	conf = flag.String("conf", "/home/tamnb/projects/src/github.com/nguyenbatam/crawl_pinterest_mysql/pinterest.toml", "config run file *.toml")
	c    = config.CrawlConfig{}
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
	for {
		listCrawlSource, err := db.GetCrawlSource()
		if err != nil {
			fmt.Println("err when connect to db", err)
			break
		}
		for _, crawlSource := range listCrawlSource {
			fmt.Println("crawlSource", "id", crawlSource.Id, "type  ", crawlSource.Type, "topic", crawlSource.TopicId)
			if crawlSource.Type == db.Keyword {
				searchedImages := GetImagesFromSearch(crawlSource.Keyword)
				fmt.Println("keyword ", crawlSource.Keyword, " len ", len(searchedImages))
				for _, imageInfo := range searchedImages {
					imageInfo.CrawledTime = time.Now().UnixNano()
					if id, _ := db.GetImageId(imageInfo.SourceId); id > 0 {
						imageInfo.Id = id
						db.InsertImageToTopic(db.ImagesInTopic{TopicId: crawlSource.TopicId, ImageId: imageInfo.Id, CrawledTime: imageInfo.CrawledTime})
						db.InsertImageToSearchKeyword(db.SearchKeyword{KeywordId: crawlSource.Id, ImageId: imageInfo.Id})
						GetAndAddRelatedPinToTopic(imageInfo, crawlSource.TopicId)
						continue
					}
					imageInfo, err = db.InsertImageInfo(imageInfo)
					if err != nil {
						fmt.Println("err when insert image info ", err, "image info ", imageInfo)
						continue
					}
					if imageInfo.Id > 0 {
						db.InsertImageToTopic(db.ImagesInTopic{TopicId: crawlSource.TopicId, ImageId: imageInfo.Id, CrawledTime: imageInfo.CrawledTime})
						db.InsertImageToSearchKeyword(db.SearchKeyword{KeywordId: crawlSource.Id, ImageId: imageInfo.Id})
						GetAndAddRelatedPinToTopic(imageInfo, crawlSource.TopicId)
					}
				}
			} else if crawlSource.Type == db.Pin {
				var imageInfo db.Image
				id, _ := db.GetImageId(crawlSource.Keyword)
				if id <= 0 {
					imageInfo, err = GetPinInfo(crawlSource.Keyword)
					if err != nil {
						fmt.Println("err when GetPinInfo ", err)
						continue
					}
					imageInfo, err = db.InsertImageInfo(imageInfo)
					if err != nil {
						fmt.Println("err when insert image info ", err, "image info ", imageInfo)
						continue
					}
				} else {
					imageInfo, err = db.GetImageInfo(id)
					if err != nil {
						fmt.Println("err get image info with id ", err)
						continue
					}
				}
				if imageInfo.Id > 0 {
					if crawlSource.ImageId == 0 {
						crawlSource.ImageId = imageInfo.Id
						err = db.UpdateImageIdInCrawlSource(crawlSource)
						if err != nil {
							fmt.Println("err when  update image id crawl source ", err)
							continue
						}
					}
					db.InsertImageToTopic(db.ImagesInTopic{TopicId: crawlSource.TopicId, ImageId: imageInfo.Id, CrawledTime: imageInfo.CrawledTime})
					relatedImages := GetAndAddRelatedPinToTopic(imageInfo, crawlSource.TopicId)
					for i := 0; i < len(relatedImages); i++ {
						GetAndAddRelatedPinToTopic(relatedImages[i], crawlSource.TopicId)
					}
				}
			}
		}
		time.Sleep(15 * time.Minute)
	}
}

func GetAndAddRelatedPinToTopic(image db.Image, topicId int) []db.Image {
	relatedImages, newSuggestedKeywords, err := GetPinFromRelatedPin(image.SourceId, c.MaxRelated)
	if err != nil {
		fmt.Println(image.SourceId)
		fmt.Println("err", err)
	}
	fmt.Println("get related pin  ", image.SourceId, "topicId", topicId, " len ", len(relatedImages), "newSuggestedKeywords", newSuggestedKeywords)
	for i := 0; i < len(newSuggestedKeywords); i++ {
		keyword := newSuggestedKeywords[i]
		crawlSource, err := db.GetCrawlSourceInfo(keyword)
		if err != nil {
			fmt.Println("err when get crawl source", err)
			continue
		}
		if crawlSource.Id <= 0 {
			crawlSource, err = db.InsertCrawlSource(db.CrawlSource{Keyword: keyword, Loop: false, Status: false, FirstCrawl: false, Type: db.Keyword})
			if err != nil {
				fmt.Println("err when insert crawl source", err)
				continue
			}
		}
		db.InsertSuggestedKeyword(db.SuggestedKeyword{KeywordId: crawlSource.Id, ImageId: image.Id})
	}

	imageInfos := make([]db.Image, 0)
	for i := 0; i < len(relatedImages); i++ {

		relatedImageInfo := relatedImages[i]
		relatedImageInfo.CrawledTime = time.Now().UnixNano()
		if id, _ := db.GetImageId(relatedImageInfo.SourceId); id > 0 {
			relatedImageInfo.Id = id
			db.InsertRelatedPinsImageInfo(db.RelatedPin{ImageId: image.Id, RelatedImageId: relatedImageInfo.Id})
			db.InsertImageToTopic(db.ImagesInTopic{TopicId: topicId, ImageId: relatedImageInfo.Id, CrawledTime: relatedImageInfo.CrawledTime})
			imageInfos = append(imageInfos, relatedImageInfo)
			continue
		}

		relatedImageInfo, err = db.InsertImageInfo(relatedImageInfo)
		if err != nil {
			fmt.Println("err when insert image info ", err, "image info ", relatedImageInfo)
			continue
		}
		if relatedImageInfo.Id > 0 {
			db.InsertImageToTopic(db.ImagesInTopic{TopicId: topicId, ImageId: relatedImageInfo.Id, CrawledTime: relatedImageInfo.CrawledTime})
			db.InsertRelatedPinsImageInfo(db.RelatedPin{ImageId: image.Id, RelatedImageId: relatedImageInfo.Id})
		}
		imageInfos = append(imageInfos, relatedImageInfo)
	}
	return imageInfos
}

func GetImagesFromSearch(keyword string) []db.Image {
	imageSearchinfos, bookmark, _ := GetPinFromKeyWordSearch(keyword)
	for j := 0; j < c.MaxPageSearch-1; j++ {
		imageNextSearchInfos, nextBookmark, err := GetPinFromNextPageSearch(keyword, bookmark)
		bookmark = nextBookmark
		if err != nil {
			fmt.Println(keyword, " page ", j, "bookmark", bookmark)
			fmt.Println("err", err)
		}
		imageSearchinfos = append(imageSearchinfos, imageNextSearchInfos...)
	}
	return imageSearchinfos
}

func GetPinFromKeyWordSearch(keyword string) ([]db.Image, string, error) {
	dataQuery := "{\"options\":{\"article\":null,\"applied_filters\":null,\"appliedProductFilters\":\"---\",\"auto_correction_disabled\":false,\"corpus\":null,\"customized_rerank_type\":null,\"filters\":null,\"query\":\"" +
		keyword +
		"\",\"query_pin_sigs\":null,\"redux_normalize_feed\":true,\"rs\":\"rs\",\"scope\":\"pins\",\"source_id\":null,\"no_fetch_context_on_resource\":false},\"context\":{}}"
	uri := "https://www.pinterest.com/resource/BaseSearchResource/get/?source_url=" + url.QueryEscape("/search/pins/?q="+url.QueryEscape(keyword)) + "&data=" + url.QueryEscape(dataQuery)
	req, err := http.NewRequest("GET", uri, nil)
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.CsrfToken})
	req.AddCookie(&http.Cookie{Name: "_pinterest_sess", Value: c.SessionKey})
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("failed to fetch bodyBytes: %d %s", resp.StatusCode, resp.Status)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	response := map[string]interface{}{}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, "", err
	}
	resourceResponse := response["resource_response"].(map[string]interface{})
	data := resourceResponse["data"].(map[string]interface{})
	results := data["results"].([]interface{})

	imageInfos := make([]db.Image, 0)
	for i := 0; i < len(results); i++ {
		info := results[i].(map[string]interface{})
		if info["type"].(string) != "pin" {
			continue
		}
		//images := info["images"].(map[string]interface{})
		//pinner := info["pinner"].(map[string]interface{})
		//imageInfo := db.Image{}
		//imageInfo.SourceId = info["id"].(string)
		//if _, err := strconv.ParseUint(imageInfo.SourceId, 10, 64); err != nil {
		//	continue
		//}
		//imageInfo.Link = "https://www.pinterest.com/pin/" + imageInfo.SourceId
		//if info["title"] != nil && len(info["title"].(string)) > 0 {
		//	imageInfo.Title = info["title"].(string)
		//} else if info["grid_title"] != nil && len(info["grid_title"].(string)) > 0 {
		//	imageInfo.Title = info["grid_title"].(string)
		//} else if info["description"] != nil && len(info["description"].(string)) > 0 {
		//	imageInfo.Title = info["description"].(string)
		//}
		//ImageData := make([]db.ImageSize, 0)
		//setUrlImage := make(map[int]bool)
		//for _typeImage, _ := range images {
		//	sizeValue := images[_typeImage].(map[string]interface{})
		//	if _typeImage == "orig" {
		//		imageInfo.Image = sizeValue["url"].(string)
		//	}
		//	imageSize := db.ImageSize{}
		//	imageSize.Url = sizeValue["url"].(string)
		//	imageSize.Width = int(sizeValue["width"].(float64))
		//	imageSize.Height = int(sizeValue["height"].(float64))
		//	if setUrlImage[imageSize.Width*imageSize.Height] {
		//		continue
		//	}
		//	setUrlImage[imageSize.Width*imageSize.Height] = true
		//	ImageData = append(ImageData, imageSize)
		//}
		//imageDataJson, err := json.Marshal(ImageData)
		//imageInfo.ImageSize = string(imageDataJson)
		//imageInfo.Link = "https://www.pinterest.com/pin/" + imageInfo.SourceId
		//createdTime := info["created_at"].(string) //Fri, 25 Sep 2020 16:51:58 +0000
		//created, err := time.Parse(time.RFC1123, createdTime)
		//if err != nil {
		//	fmt.Println("err when parser time", createdTime)
		//	continue
		//}
		//imageInfo.CreatedTime = created.UnixNano()
		//imageInfo.OwnerName = pinner["full_name"].(string)
		//imageInfo.OwnerUrl = "https://www.pinterest.com/" + pinner["username"].(string)
		//if info["board"] != nil {
		//	board := info["board"].(map[string]interface{})
		//	imageInfo.BoardName = board["name"].(string)
		//	imageInfo.BoardUrl = "https://www.pinterest.com" + board["url"].(string)
		//} else {
		//	imageInfo.BoardName = ""
		//	imageInfo.BoardUrl = ""
		//}
		imageInfo, err := ParseImageInfoFromObject(info)
		if err != nil {
			continue
		}
		imageInfos = append(imageInfos, imageInfo)

	}
	resource := response["resource"].(map[string]interface{})
	options := resource["options"].(map[string]interface{})
	bookmarks := options["bookmarks"].([]interface{})
	bookmark := ""
	for _, value := range bookmarks {
		bookmark = value.(string)
		break
	}
	return imageInfos, bookmark, nil
}

func GetPinFromRelatedPin(pinID string, size int) ([]db.Image, []string, error) {
	//https://www.pinterest.com/resource/RelatedPinFeedResource/get/?source_url=%2Fpin%2F490259109445238040%2F&data=%7B%22options%22%3A%7B%22field_set_key%22%3A%22unauth_react%22%2C%22page_size%22%3A50%2C%22pin%22%3A%22490259109445238040%22%2C%22prepend%22%3Afalse%2C%22add_vase%22%3Atrue%2C%22show_seo_canonical_pins%22%3Atrue%2C%22source%22%3A%22unknown%22%2C%22top_level_source%22%3A%22unknown%22%2C%22top_level_source_depth%22%3A1%7D%2C%22context%22%3A%7B%7D%7D
	dataQuery := "{\"options\":{\"pin_id\":\"" +
		pinID +
		"\",\"context_pin_ids\":[],\"page_size\":" +
		strconv.Itoa(size) +
		",\"search_query\":\"\",\"source\":\"deep_linking\",\"top_level_source\":\"deep_linking\",\"top_level_source_depth\":1,\"is_pdp\":false,\"no_fetch_context_on_resource\":false},\"context\":{}}"
	query := url.QueryEscape("/pin/"+pinID+"/") + "&data=" + url.QueryEscape(dataQuery)
	uri := "https://www.pinterest.com/resource/RelatedModulesResource/get/?source_url=" + query
	req, err := http.NewRequest("GET", uri, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.CsrfToken})
	req.AddCookie(&http.Cookie{Name: "_pinterest_sess", Value: c.SessionKey})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("failed to fetch bodyBytes: %d %s", resp.StatusCode, resp.Status)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	response := map[string]interface{}{}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, nil, err
	}
	resourceResponse := response["resource_response"].(map[string]interface{})
	data := resourceResponse["data"].([]interface{})
	imageInfos := make([]db.Image, 0)
	suggestedKeywords := make([]string, 0)
	for i := 0; i < len(data); i++ {
		info := data[i].(map[string]interface{})
		if info["type"].(string) != "pin" {
			objects := info["objects"].([]interface{})
			for i := 0; i < len(objects); i++ {
				object := objects[i].(map[string]interface{})
				if object["type"].(string) == "explorearticle" {
					suggestedTitle := object["title"].(map[string]interface{})
					if suggestedTitle["format"] != nil && len(suggestedTitle["format"].(string)) > 0 {
						suggestedKeyword := suggestedTitle["format"].(string)
						suggestedKeyword = strings.ToLower(suggestedKeyword)
						suggestedKeyword = strings.TrimSpace(suggestedKeyword)
						suggestedKeywords = append(suggestedKeywords, suggestedKeyword)
					}
				}
			}
			continue
		}
		//images := info["images"].(map[string]interface{})
		//pinner := info["pinner"].(map[string]interface{})
		//imageInfo := db.Image{}
		//imageInfo.SourceId = info["id"].(string)
		//if _, err := strconv.ParseUint(imageInfo.SourceId, 10, 64); err != nil {
		//	continue
		//}
		//imageInfo.Link = "https://www.pinterest.com/pin/" + imageInfo.SourceId
		//if info["title"] != nil && len(info["title"].(string)) > 0 {
		//	imageInfo.Title = info["title"].(string)
		//} else if info["grid_title"] != nil && len(info["grid_title"].(string)) > 0 {
		//	imageInfo.Title = info["grid_title"].(string)
		//} else if info["description"] != nil && len(info["description"].(string)) > 0 {
		//	imageInfo.Title = info["description"].(string)
		//}
		//ImageData := make([]db.ImageSize, 0)
		//setUrlImage := make(map[int]bool)
		//for _typeImage, _ := range images {
		//	sizeValue := images[_typeImage].(map[string]interface{})
		//	if _typeImage == "orig" {
		//		imageInfo.Image = sizeValue["url"].(string)
		//	}
		//	imageSize := db.ImageSize{}
		//	imageSize.Url = sizeValue["url"].(string)
		//	imageSize.Width = int(sizeValue["width"].(float64))
		//	imageSize.Height = int(sizeValue["height"].(float64))
		//	if setUrlImage[imageSize.Width*imageSize.Height] {
		//		continue
		//	}
		//	setUrlImage[imageSize.Width*imageSize.Height] = true
		//	ImageData = append(ImageData, imageSize)
		//}
		//imageDataJson, err := json.Marshal(ImageData)
		//imageInfo.ImageSize = string(imageDataJson)
		//createdTime := info["created_at"].(string) //Fri, 25 Sep 2020 16:51:58 +0000
		//created, err := time.Parse(time.RFC1123, createdTime)
		//if err != nil {
		//	fmt.Println("err when parser time", createdTime)
		//	continue
		//}
		//imageInfo.CreatedTime = created.UnixNano()
		//imageInfo.OwnerName = pinner["full_name"].(string)
		//imageInfo.OwnerUrl = "https://www.pinterest.com/" + pinner["username"].(string)
		//if info["board"] != nil {
		//	board := info["board"].(map[string]interface{})
		//	imageInfo.BoardName = board["name"].(string)
		//	imageInfo.BoardUrl = "https://www.pinterest.com" + board["url"].(string)
		//} else {
		//	imageInfo.BoardName = ""
		//	imageInfo.BoardUrl = ""
		//}
		imageInfo, err := ParseImageInfoFromObject(info)
		if err != nil {
			continue
		}
		imageInfos = append(imageInfos, imageInfo)
	}
	return imageInfos, suggestedKeywords, nil
}

func GetPinFromNextPageSearch(keyword string, bookmark string) ([]db.Image, string, error) {
	bodyPost := []byte("source_url=" + url.QueryEscape("/search/pins/?q="+url.QueryEscape(keyword)) + "&data=" +
		url.QueryEscape("{\"options\":{\"article\":null,\"applied_filters\":null,\"appliedProductFilters\":\"---\",\"auto_correction_disabled\":false,\"corpus\":null,\"customized_rerank_type\":null,\"filters\":null,\"query\":\""+
			keyword+
			"\",\"query_pin_sigs\":null,\"redux_normalize_feed\":true,\"rs\":\"typed\",\"scope\":\"pins\",\"source_id\":null,\"bookmarks\":[\""+
			bookmark+
			"\"],\"no_fetch_context_on_resource\":false},\"context\":{}}"))
	req, err := http.NewRequest("POST", "https://www.pinterest.com/resource/BaseSearchResource/get/", bytes.NewBuffer(bodyPost))
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.CsrfToken})
	req.AddCookie(&http.Cookie{Name: "_pinterest_sess", Value: c.SessionKey})
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")
	req.Header.Set("x-csrftoken", c.CsrfToken)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("failed to fetch bodyBytes: %d %s", resp.StatusCode, resp.Status)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	response := map[string]interface{}{}
	err = json.Unmarshal(bodyBytes, &response)

	if err != nil {
		//fmt.Println(err)
		return nil, "", err
	}
	resourceResponse := response["resource_response"].(map[string]interface{})
	data := resourceResponse["data"].(map[string]interface{})
	results := data["results"].([]interface{})
	imageInfos := make([]db.Image, 0)
	for i := 0; i < len(results); i++ {
		info := results[i].(map[string]interface{})
		if info["type"].(string) != "pin" {
			continue
		}
		//images := info["images"].(map[string]interface{})
		//pinner := info["pinner"].(map[string]interface{})
		//imageInfo := db.Image{}
		//imageInfo.SourceId = info["id"].(string)
		//if _, err := strconv.ParseUint(imageInfo.SourceId, 10, 64); err != nil {
		//	continue
		//}
		//imageInfo.Link = "https://www.pinterest.com/pin/" + imageInfo.SourceId
		//if info["title"] != nil && len(info["title"].(string)) > 0 {
		//	imageInfo.Title = info["title"].(string)
		//} else if info["grid_title"] != nil && len(info["grid_title"].(string)) > 0 {
		//	imageInfo.Title = info["grid_title"].(string)
		//} else if info["description"] != nil && len(info["description"].(string)) > 0 {
		//	imageInfo.Title = info["description"].(string)
		//}
		//ImageData := make([]db.ImageSize, 0)
		//setUrlImage := make(map[int]bool)
		//for _typeImage, _ := range images {
		//	sizeValue := images[_typeImage].(map[string]interface{})
		//	if _typeImage == "orig" {
		//		imageInfo.Image = sizeValue["url"].(string)
		//	}
		//	imageSize := db.ImageSize{}
		//	imageSize.Url = sizeValue["url"].(string)
		//	imageSize.Width = int(sizeValue["width"].(float64))
		//	imageSize.Height = int(sizeValue["height"].(float64))
		//	if setUrlImage[imageSize.Width*imageSize.Height] {
		//		continue
		//	}
		//	setUrlImage[imageSize.Width*imageSize.Height] = true
		//	ImageData = append(ImageData, imageSize)
		//}
		//imageDataJson, err := json.Marshal(ImageData)
		//imageInfo.ImageSize = string(imageDataJson)
		//createdTime := info["created_at"].(string) //Fri, 25 Sep 2020 16:51:58 +0000
		//created, err := time.Parse(time.RFC1123, createdTime)
		//if err != nil {
		//	fmt.Println("err when parser time", createdTime)
		//	continue
		//}
		//imageInfo.CreatedTime = created.UnixNano()
		//imageInfo.OwnerName = pinner["full_name"].(string)
		//imageInfo.OwnerUrl = "https://www.pinterest.com/" + pinner["username"].(string)
		//if info["board"] != nil {
		//	board := info["board"].(map[string]interface{})
		//	imageInfo.BoardName = board["name"].(string)
		//	imageInfo.BoardUrl = "https://www.pinterest.com" + board["url"].(string)
		//} else {
		//	imageInfo.BoardName = ""
		//	imageInfo.BoardUrl = ""
		//}
		imageInfo, err := ParseImageInfoFromObject(info)
		if err != nil {
			continue
		}
		imageInfos = append(imageInfos, imageInfo)
		//imageInfos = append(imageInfos, imageInfo)
	}

	resource := response["resource"].(map[string]interface{})
	options := resource["options"].(map[string]interface{})
	bookmarks := options["bookmarks"].([]interface{})
	nextBookmark := ""
	for _, value := range bookmarks {
		nextBookmark = value.(string)
		break
	}
	return imageInfos, nextBookmark, nil
}

func GetPinInfo(pinID string) (db.Image, error) {
	image := db.Image{}
	resp, err := http.Get("https://www.pinterest.com/pin/" + pinID)
	if err != nil {
		return image, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return image, fmt.Errorf("failed to fetch data: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return image, err
	}
	data := doc.Find("#__PWS_DATA__").Text()

	ioutil.WriteFile("1.txt", []byte(data), fs.ModePerm)
	response := map[string]interface{}{}
	err = json.Unmarshal([]byte(data), &response)
	if err != nil {
		return image, err
	}
	props := response["props"].(map[string]interface{})
	initialReduxState := props["initialReduxState"].(map[string]interface{})
	pins := initialReduxState["pins"].(map[string]interface{})
	if pins[pinID] == nil {
		return image, fmt.Errorf("can request data with Pin ID :%v", pinID)
	}

	info := pins[pinID].(map[string]interface{})

	images := info["images"].(map[string]interface{})

	image.SourceId = info["id"].(string)
	if _, err := strconv.ParseUint(image.SourceId, 10, 64); err != nil {
		return image, err
	}
	image.Link = "https://www.pinterest.com/pin/" + image.SourceId
	ImageData := make([]db.ImageSize, 0)
	setUrlImage := make(map[int]bool)
	for _typeImage, _ := range images {
		sizeValue := images[_typeImage].(map[string]interface{})
		if _typeImage == "orig" {
			image.Image = sizeValue["url"].(string)
		}
		imageSize := db.ImageSize{}
		imageSize.Url = sizeValue["url"].(string)
		imageSize.Width = int(sizeValue["width"].(float64))
		imageSize.Height = int(sizeValue["height"].(float64))
		if setUrlImage[imageSize.Width*imageSize.Height] {
			continue
		}
		setUrlImage[imageSize.Width*imageSize.Height] = true
		ImageData = append(ImageData, imageSize)
	}
	imageDataJson, err := json.Marshal(ImageData)
	image.ImageSize = string(imageDataJson)
	createdTime := info["created_at"].(string) //Fri, 25 Sep 2020 16:51:58 +0000
	created, err := time.Parse(time.RFC1123, createdTime)
	if err != nil {
		fmt.Println("err when parser time", createdTime)
		return image, err
	}
	image.CreatedTime = created.UnixNano()

	image.Title = ""
	if info["title"] != nil {
		image.Title = strings.TrimSpace(info["title"].(string))
	}
	if len(image.Title) == 0 {
		if info["grid_title"] != nil {
			gridTitle := strings.TrimSpace(info["grid_title"].(string))
			if len(gridTitle) > 0 {
				image.Title = gridTitle
			}
		}

	}
	image.Description = ""
	if info["description"] != nil {
		image.Description = strings.TrimSpace(info["description"].(string))
	}

	closeupUnifiedDescription := strings.TrimSpace(info["closeup_unified_description"].(string))
	if len(image.Title) == 0 {
		if len(image.Description) > 0 {
			image.Title = image.Description
		} else if len(closeupUnifiedDescription) > 0 {
			image.Title = closeupUnifiedDescription
		} else {
			image.Title = ""
		}
	}
	if len(image.Description) == 0 {
		if len(closeupUnifiedDescription) > 0 {
			image.Description = closeupUnifiedDescription
		} else if len(info["description_html"].(string)) > 0 {
			image.Description = strings.TrimSpace(info["description_html"].(string))
		} else {
			image.Description = ""
		}
	}

	var closeupAttribution map[string]interface{}

	if info["closeup_attribution"] != nil {
		closeupAttribution = info["closeup_attribution"].(map[string]interface{})
	} else {
		closeupAttribution = info["pinner"].(map[string]interface{})
	}

	image.OwnerName = closeupAttribution["full_name"].(string)
	image.OwnerUrl = "https://www.pinterest.com/" + closeupAttribution["username"].(string)

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
	hashtags := make([]string, 0)
	if info["hashtags"] != nil {
		hashtagsInfo := info["hashtags"].([]interface{})
		for _, key := range hashtagsInfo {
			hashtags = append(hashtags, key.(string))
		}
	}
	boards := initialReduxState["boards"].(map[string]interface{})
	image.BoardDescription = ""
	for _, v := range boards {
		board := v.(map[string]interface{})
		if board["description"] != nil {
			image.BoardDescription = board["description"].(string)
		}
		break
	}
	return image, nil
}

func ParseImageInfoFromObject(info map[string]interface{}) (db.Image, error) {
	imageInfo := db.Image{}
	images := info["images"].(map[string]interface{})
	pinner := info["pinner"].(map[string]interface{})

	imageInfo.SourceId = info["id"].(string)
	if _, err := strconv.ParseUint(imageInfo.SourceId, 10, 64); err != nil {
		return imageInfo, err
	}
	imageInfo.Link = "https://www.pinterest.com/pin/" + imageInfo.SourceId
	if info["title"] != nil && len(info["title"].(string)) > 0 {
		imageInfo.Title = info["title"].(string)
	} else if info["grid_title"] != nil && len(info["grid_title"].(string)) > 0 {
		imageInfo.Title = info["grid_title"].(string)
	} else if info["description"] != nil && len(info["description"].(string)) > 0 {
		imageInfo.Title = info["description"].(string)
	}
	ImageData := make([]db.ImageSize, 0)
	setUrlImage := make(map[int]bool)
	for _typeImage, _ := range images {
		sizeValue := images[_typeImage].(map[string]interface{})
		if _typeImage == "orig" {
			imageInfo.Image = sizeValue["url"].(string)
		}
		imageSize := db.ImageSize{}
		imageSize.Url = sizeValue["url"].(string)
		imageSize.Width = int(sizeValue["width"].(float64))
		imageSize.Height = int(sizeValue["height"].(float64))
		if setUrlImage[imageSize.Width*imageSize.Height] {
			continue
		}
		setUrlImage[imageSize.Width*imageSize.Height] = true
		ImageData = append(ImageData, imageSize)
	}
	imageDataJson, err := json.Marshal(ImageData)
	imageInfo.ImageSize = string(imageDataJson)
	createdTime := info["created_at"].(string) //Fri, 25 Sep 2020 16:51:58 +0000
	created, err := time.Parse(time.RFC1123, createdTime)
	if err != nil {
		fmt.Println("err when parser time", createdTime)
		return imageInfo, err
	}
	imageInfo.CreatedTime = created.UnixNano()
	imageInfo.OwnerName = pinner["full_name"].(string)
	imageInfo.OwnerUrl = "https://www.pinterest.com/" + pinner["username"].(string)
	if info["board"] != nil {
		board := info["board"].(map[string]interface{})
		imageInfo.BoardName = board["name"].(string)
		imageInfo.BoardUrl = "https://www.pinterest.com" + board["url"].(string)
	} else {
		imageInfo.BoardName = ""
		imageInfo.BoardUrl = ""
	}
	return imageInfo, nil
}

//
//
//
//source_url=%2Fsearch%2Fpins%2F%3Fq%3Dharry%2Bpotter&
//	data=%7B%22options%22%3A%7B%22article%22%3Anull%2C%22applied_filters%22%3Anull%2C%22appliedProductFilters%22%3A%22---%22%2C%22auto_correction_disabled%22%3Afalse%2C%22corpus%22%3Anull%2C%22customized_rerank_type%22%3Anull%2C%22filters%22%3Anull%2C%22query%22%3A%22the+oscars+winners+portraits%22%2C%22query_pin_sigs%22%3Anull%2C%22redux_normalize_feed%22%3Atrue%2C%22rs%22%3A%22typed%22%2C%22scope%22%3A%22pins%22%2C%22source_id%22%3Anull%2C%22bookmarks%22%3A%5B%22Y2JVSG81V2sxcmNHRlpWM1J5VFVaU1YxWllhRlJXTVVreVZsZDRRMVV4U1hsVlZFWlhVbnBXTTFWNlNrZFdNa3BIWVVaYVdGSXlhRkZXUm1Rd1ZtMVJlRlZ1VWs1V1ZGWlBXVmh3UjFac1draE5WRkpWVFd0YWVWUnNhRXRYUmxsNlVXMUdWVlpzY0hsYVZscExWbFpXZEZKc1RsTmlXRkV4Vm10U1IxVXhUbkpPVlZwUVZsWmFXVmxzYUZOVU1YQllaRVYwYWxKc1NubFhhMVpoWWtaYVZWWnVhRmRpUmtwVVYxWmFTbVF3TVZWV2JGWk9VbXR3U1ZkWGVGWmxSVFYwVW10b2FGSnJTbGhWYlhoYVRXeFplVTFZWkZOaGVrWjVWR3hhVjFadFNsVlNibEpXWWtaS1dGVnFSbUZqVmxKeFZHeEdWbFpFUVRWYWExcFhVMWRLTmxWdGVGZE5XRUpLVm10amVHSXhiRmRUV0dob1RUSlNXVmxVUmt0VU1WSlhWbGhvYWxZd2NFbFpWVlUxWVVkRmVGZFljRmRTTTFKeVZqSXhWMk15VGtkV2JFcFlVakpvYUZadGRHRlpWMDVYVlc1S1ZtSnJOVzlVVm1RMFpVWldjMVZzVGxWaVJYQkpXWHBPYzFaV1dsZFRia1poVmpOTmVGVnNXbE5YVjBwR1QxWmtVMDFFUWpOV2FrbDRaREZrZEZac1pHbFRSa3BXVm10Vk1XRkdiRmhOVjNCc1lrWktlbGRyV21GVWJFcDFVVzVvVmxac1NraFdSRVpMVWpKS1JWZHNhRmRpUlhCRVYyeGFWbVZIVWtkVGJrWm9VbXhhYjFSV1duZFhiR1IwWkVWYVVGWnJTbE5WUmxGNFQwVXhObFJVVWs1bGJHdDVWRmN4U2sxR2NFVlhiV3hoWWxWcmQxUnJVbkpsVm14WVVtMXNXbFpGVmpOWFYzQnVUVEZyZVZKWWNFOVdSVEF3VjJ4U1FrMVdiRFpWVkU1aFlsWktjRmRyWkU5aVJteHhWbGh3V2sxc1NtOVVWbVJHWkRBeFNGTnRNVTVoYkhCdlZGWmtWazFzYTNwbFJUbFRWbTFSTkdaSFZteFBSRWt3VG1wRk1FNVVVVE5OVkUxNFRWZEplRTVFVm0xYWJVWnFXWHBWTkU5WFdtbFpWRTAwVFVkR2FVOUVVbTFOVkdoc1RtcEZlRTFxUlRKT01rbDVXbXBzYUZscVNYaE9SRlUwVG5wYWFFMTZaRGhVYTFaWVprRTlQUT09fFVIbzVUMkl5Tld4bVJFNXJUMVJOTUZwVVFUSlBSR3MxVFRKUmVFNHlWVFZhUkVVelQwZE9iRmxVU1RSWlYxcHRUbFJuTlU1NlNYaGFhbHBvVG5wc2JFNUVRWGhhVkVGNVdrZFpNazR5UlRSTmVtaHFXVlJhYTA5WFVYaGFSRm80Vkd0V1dHWkJQVDA9fGYwYTc0N2I5MjU2MmQyZDA2YmQ2NGM3Y2MxMjVjOTA5YWY4MjAwMmJkODgyNzlhYTFhMjliNTJiYmU4ZGZmN2J8TkVXfA%3D%3D%22%5D%2C%22no_fetch_context_on_resource%22%3Afalse%7D%2C%22context%22%3A%7B%7D%7D
//source_url=%2Fsearch%2Fpins%2F%3Fq%3Dthe%2520oscars%2520winners%2520portraits%26rs%3Dtyped%26term_meta%5B%5D%3Dthe%2520oscars%2520winners%2520portraits%257Ctyped&
//	data=%7B%22options%22%3A%7B%22article%22%3Anull%2C%22applied_filters%22%3Anull%2C%22appliedProductFilters%22%3A%22---%22%2C%22auto_correction_disabled%22%3Afalse%2C%22corpus%22%3Anull%2C%22customized_rerank_type%22%3Anull%2C%22filters%22%3Anull%2C%22query%22%3A%22the%20oscars%20winners%20portraits%22%2C%22query_pin_sigs%22%3Anull%2C%22redux_normalize_feed%22%3Atrue%2C%22rs%22%3A%22typed%22%2C%22scope%22%3A%22pins%22%2C%22source_id%22%3Anull%2C%22bookmarks%22%3A%5B%22Y2JVSG81V2sxcmNHRlpWM1J5VFVaU1YxWllhRlJXTVVreVZsZDRRMVV4U1hsVlZFWlhVbnBXTTFWNlNrZFdNa3BIWVVaYVdGSXlhRkZXUm1Rd1ZtMVJlRlZ1VWs1V1ZGWlBXVmh3UjFac1draE5WRkpWVFd0YWVWUnNhRXRYUmxsNlVXMUdWVlpzY0hsYVZscExWbFpXZEZKc1RsTmlXRkV4Vm10U1IxVXhUbkpPVlZwUVZsWmFXVmxzYUZOVU1YQllaRVYwYWxKc1NubFhhMVpoWWtaYVZWWnVhRmRpUmtwVVYxWmFTbVF3TVZWV2JGWk9VbXR3U1ZkWGVGWmxSVFYwVW10b2FGSnJTbGhWYlhoYVRXeFplVTFZWkZOaGVrWjVWR3hhVjFadFNsVlNibEpXWWtaS1dGVnFSbUZqVmxKeFZHeEdWbFpFUVRWYWExcFhVMWRLTmxWdGVGZE5XRUpLVm10amVHSXhiRmRUV0dob1RUSlNXVmxVUmt0VU1WSlhWbGhvYWxZd2NFbFpWVlUxWVVkRmVGZFljRmRTTTFKeVZqSXhWMk15VGtkV2JFcFlVakpvYUZadGRHRlpWMDVYVlc1S1ZtSnJOVzlVVm1RMFpVWldjMVZzVGxWaVJYQkpXWHBPYzFaV1dsZFRia1poVmpOTmVGVnNXbE5YVjBwR1QxWmtVMDFFUWpOV2FrbDRaREZrZEZac1pHbFRSa3BXVm10Vk1XRkdiRmhOVjNCc1lrWktlbGRyV21GVWJFcDFVVzVvVmxac1NraFdSRVpMVWpKS1JWZHNhRmRpUlhCRVYyeGFWbVZIVWtkVGJrWm9VbXhhYjFSV1duZFhiR1IwWkVWYVVGWnJTbE5WUmxGNFQwVXhObFJVVWs1bGJHdDVWRmN4U2sxR2NFVlhiV3hoWWxWcmQxUnJVbkpsVm14WVVtMXNXbFpGVmpOWFYzQnVUVEZyZVZKWWNFOVdSVEF3VjJ4U1FrMVdiRFpWVkU1aFlsWktjRmRyWkU5aVJteHhWbGh3V2sxc1NtOVVWbVJHWkRBeFNGTnRNVTVoYkhCdlZGWmtWazFzYTNwbFJUbFRWbTFSTkdaSFZteFBSRWt3VG1wRk1FNVVVVE5OVkUxNFRWZEplRTVFVm0xYWJVWnFXWHBWTkU5WFdtbFpWRTAwVFVkR2FVOUVVbTFOVkdoc1RtcEZlRTFxUlRKT01rbDVXbXBzYUZscVNYaE9SRlUwVG5wYWFFMTZaRGhVYTFaWVprRTlQUT09fFVIbzVUMkl5Tld4bVJFNXJUMVJOTUZwVVFUSlBSR3MxVFRKUmVFNHlWVFZhUkVVelQwZE9iRmxVU1RSWlYxcHRUbFJuTlU1NlNYaGFhbHBvVG5wc2JFNUVRWGhhVkVGNVdrZFpNazR5UlRSTmVtaHFXVlJhYTA5WFVYaGFSRm80Vkd0V1dHWkJQVDA9fGYwYTc0N2I5MjU2MmQyZDA2YmQ2NGM3Y2MxMjVjOTA5YWY4MjAwMmJkODgyNzlhYTFhMjliNTJiYmU4ZGZmN2J8TkVXfA%3D%3D%22%5D%2C%22no_fetch_context_on_resource%22%3Afalse%7D%2C%22context%22%3A%7B%7D%7D' \
