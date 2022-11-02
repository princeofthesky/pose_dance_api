package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"io/ioutil"
	"math"
	"os"
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
	mysqlDb := db.GetDb()
	imageIds := []db.CrawlSource{}
	err = mysqlDb.Model((*db.CrawlSource)(nil)).Column("*").Order("keyword ASC").ForEach(
		func(c *db.CrawlSource) error {
			imageIds = append(imageIds, *c)
			return nil
		})
	if err != nil {
		fmt.Println("err when find source id NULL", err)
	}
	total := map[string]map[int]bool{}
	for i := 0; i < len(imageIds); i++ {
		keyword := imageIds[i].Keyword
		ids := total[keyword]
		if ids == nil {
			ids = map[int]bool{}
			total[keyword] = ids
		}
		total[keyword][imageIds[i].Id] = true
	}
	fmt.Println("total ", len(total))
	totalDelete := 0
	result,err:=mysqlDb.Model((*db.SuggestedKeyword)(nil)).Where("keyword_id is NULL ").Delete()
	fmt.Println("delete NULL SuggestedKeyword ",result,err)
	if err !=nil{
		os.Exit(0)
	}
	result,err=mysqlDb.Model((*db.SuggestedKeyword)(nil)).Where("image_id is NULL ").Delete()
	fmt.Println("delete NULL SuggestedKeyword ",result,err)
	if err !=nil{
		os.Exit(0)
	}
	for _, keywords := range total {
		if len(keywords) > 1 {
			min := math.MaxInt
			for id, _ := range keywords {
				if id < min {
					min = id
				}
			}
			delete(keywords, min)
			for keywordId := range keywords {
				mysqlDb.Model((*db.SuggestedKeyword)(nil)).Where("keyword_id = ? ", keywordId).ForEach(
					func(c *db.SuggestedKeyword) error {
						c.KeywordId = min
						_, err := mysqlDb.Model(c).Where("keyword_id = ? and image_id =?", keywordId, c.ImageId).Column("keyword_id").Update()
						if err != nil {
							fmt.Println(err)
						}
						return nil
					})
				result,err:=mysqlDb.Model((*db.SuggestedKeyword)(nil)).Where("keyword_id = ? ", keywordId).Delete()
				if err !=nil{
					fmt.Println("delete loop SuggestedKeyword ",result,err)
					os.Exit(0)
				}
				result,err=mysqlDb.Model((*db.CrawlSource)(nil)).Where("id = ? ", keywordId).Delete()
				if err !=nil{
					fmt.Println("delete loop CrawlSource ",result,err)
					os.Exit(0)
				}
				fmt.Println(keywordId)
				totalDelete++
			}
		}
	}
	mysqlDb.Close()
	fmt.Println("totalDelete", totalDelete)
}
