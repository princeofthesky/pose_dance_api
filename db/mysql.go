package db

import (
	"context"
	"github.com/go-pg/pg/v10"
	"go-pinterest/config"
)

var mysqlDb *pg.DB

func Init(postgres config.Postgres) error {
	mysqlDb = pg.Connect(&pg.Options{
		Addr:     postgres.Addr,
		User:     postgres.User,
		Password: postgres.Password,
		Database: postgres.Database,
	})
	return mysqlDb.Ping(context.Background())
}
func Close() {
	mysqlDb.Close()
}
func GetDb() *pg.DB {
	return mysqlDb
}

func GetCrawlSource() ([]CrawlSource, error) {
	listCrawl := make([]CrawlSource, 0)
	err := mysqlDb.Model((*CrawlSource)(nil)).Column("*").Order("id DESC").ForEach(
		func(c *CrawlSource) error {
			if !c.Status {
				return nil
			}
			if c.Loop {
				listCrawl = append(listCrawl, *c)
				return nil
			}
			if !c.FirstCrawl {
				listCrawl = append(listCrawl, *c)
				return nil
			}
			return nil
		})

	return listCrawl, err
}

func UpdateImageIdInCrawlSource(c CrawlSource) error {
	_, err := mysqlDb.Model(&c).WherePK().Column("image_id").Update()
	return err
}
func GetImageId(sourceId string) (int, error) {
	imageInfo := Image{Id: 0, SourceId: sourceId}
	err := mysqlDb.Model(&imageInfo).Where("\"image\".\"source_id\"=?", imageInfo.SourceId).Select()
	return imageInfo.Id, err
}

func GetImageInfo(imageId int) (Image, error) {
	imageInfo := Image{Id: imageId}
	err := mysqlDb.Model(&imageInfo).WherePK().Select()
	return imageInfo, err
}

func InsertImageInfo(info Image) (Image, error) {
	_, err := mysqlDb.Model(&info).Insert()
	return info, err
}

func UpdateImageInfo(info Image) error {
	_, err := mysqlDb.Model(&info).WherePK().Column("title", "description", "owner_name", "owner_url", "board_description").Update()
	return err
}

func InsertAnnotationImage(info Annotation) error {
	_, err := mysqlDb.Model(&info).Insert()
	return err
}

func InsertImageToTopic(imageTopic ImagesInTopic) error {
	_, err := mysqlDb.Model(&imageTopic).Insert()
	return err
}

func InsertImageToSearchKeyword(searchInfo SearchKeyword) error {
	_, err := mysqlDb.Model(&searchInfo).Insert()
	return err
}

func InsertRelatedPinsImageInfo(related RelatedPin) error {
	_, err := mysqlDb.Model(&related).Insert()
	return err
}

func InsertCrawlSource(crawlSource CrawlSource) (CrawlSource, error) {
	_, err := mysqlDb.Model(&crawlSource).Insert()
	return crawlSource, err
}

func GetCrawlSourceInfo(keyword string) (CrawlSource, error) {
	c := CrawlSource{Keyword: keyword}
	err := mysqlDb.Model(&c).Where("keyword = ?", c.Keyword).Select()
	return c, err
}

func UpdateCrawlSource(crawlSource CrawlSource) (CrawlSource, error) {
	_, err := mysqlDb.Model(&crawlSource).WherePK().Column("status", "loop", "first_crawl", "topic_id").Update()
	return crawlSource, err
}

func InsertSuggestedKeyword(suggestedKeyword SuggestedKeyword) error {
	_, err := mysqlDb.Model(&suggestedKeyword).Insert()
	return err
}
