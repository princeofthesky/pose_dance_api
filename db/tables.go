package db

const (
	Keyword = 1
	Pin     = 2
)

type ImageSize struct {
	Url    string
	Width  int
	Height int
}
type Image struct {
	Image            string
	Id               int
	Title            string
	Description      string
	SourceId         string
	Link             string
	OwnerName        string
	OwnerUrl         string
	BoardName        string
	BoardUrl         string
	BoardDescription string
	ImageSize        string
	CreatedTime      int64
	CrawledTime      int64
}
type ImageGif struct {
	ImageId int
	Url     string
	Width   int
	Height  int
}
type Topic struct {
	Name   string
	Id     int
	Hidden bool
}

type ImagesInTopic struct {
	ImageId     int
	TopicId     int
	CrawledTime int64
}

type Annotation struct {
	ImageId int
	Data    string
}

type RelatedPin struct {
	ImageId        int
	RelatedImageId int
}

type SearchKeyword struct {
	ImageId   int
	KeywordId int
}

type SuggestedKeyword struct {
	ImageId   int
	KeywordId int
}

type CrawlSource struct {
	Keyword    string
	Status     bool
	Loop       bool
	FirstCrawl bool
	TopicId    int
	Type       int
	ImageId    int
	Id         int
}
