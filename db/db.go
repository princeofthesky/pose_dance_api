package db

import "go-pinterest/config"

func Init(config config.CrawlConfig) error{
	err:=initMysql(config.Postgres)
	if err !=nil {
		return  err
	}
	return initRedis(config.Redis)
}
func Close()  {
	closeMysql()
	closeRedis()
}
