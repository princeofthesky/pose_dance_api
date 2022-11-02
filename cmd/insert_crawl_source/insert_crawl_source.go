package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var (
	source = flag.String("source", "GIF_source.csv", "source category Pinterest")
	conf   = flag.String("conf", "pinterest.toml", "config run file *.toml")
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

	// Open CSV file
	fileContent, err := os.Open(*source)
	if err != nil {
		fmt.Println("err", err)
	}
	defer fileContent.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(fileContent).ReadAll()
	if err != nil {
		fmt.Println("err", err)
	}
	mysql := db.GetDb()
	oldTopicName := ""
	for i := 1; i < len(lines); i++ {
		dataLine := lines[i]
		topicName := dataLine[1]
		topicName = strings.TrimSpace(topicName)
		topicName = strings.Trim(topicName, "\n")
		if len(topicName) == 0 {
			topicName = oldTopicName
			if len(topicName) == 0 {
				continue
			}
		}
		topic := db.Topic{Name: topicName}
		err = mysql.Model(&topic).Where("\"topic\".\"name\"=?", topicName).Select()
		if err != nil {
			fmt.Println("err find topic", err)
		}
		if topic.Id <= 0 {
			mysql.Model(&topic).Insert()
		}
		fmt.Println(topic)
		keywordLines := strings.Split(dataLine[2], "\n")
		for j := 0; j < len(keywordLines); j++ {
			keywordLine := strings.TrimSpace(keywordLines[j])
			keywords := strings.Split(keywordLine, ",")
			for k := 0; k < len(keywords); k++ {
				keyword := strings.TrimSpace(keywords[k])
				if len(keyword) > 0 {
					fmt.Println(keyword, topicName)
					crawlSource := db.CrawlSource{Keyword: keyword, Status: true, Loop: true, FirstCrawl: false, Type: db.Keyword, TopicId: topic.Id}
					err = mysql.Model(&crawlSource).Where("\"crawl_source\".\"keyword\"=?", keyword).Select()
					if err != nil {
						fmt.Println("err when find keyword", err)
					}
					if crawlSource.Id == 0 {
						crawlSource, err = db.InsertCrawlSource(crawlSource)
					} else {
						crawlSource.Status = true
						crawlSource.Loop = true
						crawlSource.FirstCrawl = false
						crawlSource.TopicId = topic.Id
						crawlSource, err = db.UpdateCrawlSource(crawlSource)
						if err != nil {
							fmt.Println("err when update keyword info ", err)
						}
						//fmt.Println("keyword exits", crawlSource)
					}
					fmt.Println("topic", topicName, "topic id ", topic.Id, "keyword", keyword, " id ", crawlSource.Id)
				}
			}
		}
		pinLines := strings.Split(dataLine[3], "\n")
		for j := 0; j < len(pinLines); j++ {
			pinLine := strings.TrimSpace(pinLines[j])
			pins := strings.Split(pinLine, ",")
			for k := 0; k < len(pins); k++ {
				pinURL := strings.TrimSpace(pins[k])
				pin := strings.Trim(pinURL, "https://www.pinterest.com/pin/")
				pin = strings.Trim(pin, "https://www.pinterest.com.au/pin/")
				pin = strings.Trim(pin, "/")
				if len(pin) > 0 {
					_, err := strconv.ParseInt(pin, 10, 64)
					if err == nil {
						crawlSource := db.CrawlSource{Keyword: pin, Status: true, Loop: true, FirstCrawl: false, Type: db.Pin, TopicId: topic.Id}
						err = mysql.Model(&crawlSource).Where("\"crawl_source\".\"keyword\"=?", pin).Select()
						//if err != nil {
						//	fmt.Println("err", err)
						//}
						if crawlSource.Id == 0 {
							crawlSource, err = db.InsertCrawlSource(crawlSource)
						} else {
							crawlSource.Status = true
							crawlSource.Loop = true
							crawlSource.FirstCrawl = false
							crawlSource.TopicId = topic.Id
							crawlSource, err = db.UpdateCrawlSource(crawlSource)
							//if err != nil {
							//	fmt.Println("err when update keyword info ", err)
							//}
							//fmt.Println("keyword exits", crawlSource)
						}
						fmt.Println("topic", topicName, "topic id ", topic.Id, "pin", pin, " id ", crawlSource.Id)
					}
				}
			}
		}
		oldTopicName=topicName
	}

}
