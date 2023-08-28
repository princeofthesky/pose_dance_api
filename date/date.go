package date

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	SECOND_PER_DAY = int64(24 * 60 * 60)
	mapMonthValue  = map[time.Month]int{
		time.January:   1,
		time.February:  2,
		time.March:     3,
		time.April:     4,
		time.May:       5,
		time.June:      6,
		time.July:      7,
		time.August:    8,
		time.September: 9,
		time.October:   10,
		time.November:  11,
		time.December:  12,
	}
)

func ParseDate(dateTime string) int64 {
	dateTime = strings.TrimSpace(dateTime)
	re := regexp.MustCompile("[^0-9]+")
	splits := re.Split(dateTime, -1)
	if len(splits) < 6 {
		return 0
	}
	year, err := strconv.Atoi(splits[0])
	if err != nil {
		return 0
	}
	month, err := strconv.Atoi(splits[1])
	if err != nil {
		return 0
	}
	day, err := strconv.Atoi(splits[2])
	if err != nil {
		return 0
	}
	hour, err := strconv.Atoi(splits[3])
	if err != nil {
		return 0
	}
	minute, err := strconv.Atoi(splits[4])
	if err != nil {
		return 0
	}
	second, err := strconv.Atoi(splits[5])
	if err != nil {
		return 0
	}
	date := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	return date.Unix()
}

func FormatDate(date time.Time) string {
	year, month, day := date.Date()
	dateTime := strconv.Itoa(year)
	if mapMonthValue[month] < 10 {
		dateTime = dateTime + "/0" + strconv.Itoa(mapMonthValue[month])
	} else {
		dateTime = dateTime + "/" + strconv.Itoa(mapMonthValue[month])
	}
	if day < 10 {
		dateTime = dateTime + "/0" + strconv.Itoa(day)
	} else {
		dateTime = dateTime + "/" + strconv.Itoa(day)
	}
	return dateTime
}

func GetFolderDaily(timeStamp int64) string {
	now := time.Unix(timeStamp, 0)
	return FormatDate(now)
}
