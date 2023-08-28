package config

type Postgres struct {
	Addr     string
	User     string
	Password string
	Database string
}

type Redis struct {
	Addr     string
	Password string
	Db       int
}

type CrawlConfig struct {
	Postgres  Postgres
	Redis     Redis
	StartDate int
	EndDate   int
}
