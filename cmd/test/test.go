package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"io/ioutil"
)

var (
	//SessionKey    = "TWc9PSZ5azdObjM3eEJJTVlJQUNjTStlSzU3ajhGWC8veUI3T1VycUVrc3Y4dmkzVlgvei9ZTW91T1JiTzJycTZHMHRJVjBDZ1hCSWMwSERSREpFUkdvNy9wQStZRGF2VktkVVhKN2V1LzU1WGpKa0hpYzFidG96UkdZQXdITWFDenZJVVZ2dVl3aVBHT1IyK3NiZnJJR0hWb1plRER5Z2hPM1RHc1BhWWVhdUlLbjg1MlkzTnVDTTVZNjFtbjBYQndQVmF0THhWVk85b29hMFV2OUx6NVRFMkFpZlAyZStaU0dCQUovQWVtQTVuR2xDNm9NaU0yL3RTdXV5aGhXVXprN0dOOU1EU3lVcy9VMUk2SVpzK0FHYUE5TExLZGFUVmY5aEdJQVJmcjJVUnBlc3ZhVENxRzFoY3A1UHFOSDlIYXdZM1RpQ2sxb2FKdXJTZ2pGMlRvNzdUZFFOK2ZlTXZVdGx0cnFFUWpuSzVtTU1GTzhWejZKcWw3all0VlM1aVRQcFdZdHpQK0Q4MGdGZUhEMnN0d0xURThOY2gwMUt6QXpDRVZJcllSQXBVZmhKMzROUTUyU21EclE1VE5Lb1E3TVlPSU9zZE5xRlNIcnJ3eXhWc2xIbTBvcnBnSHlLMlBYQnEwYm5kQSt4dXVMK0VTTmVqMWpGQThzUjdiTDdNU1lXVFhWaVpxQ3hEK3E0R2ZWZlo3ek1selBWbDJiN3VnMlJrRE1DTXRYeXJDeGhWK0dqeGJqRWVWNTdBNW9ZcWkvdmRQU0NNMmszZlNKaGtBSG1icFExdkF6Mjh4TDFydXFyNk5GZmZTSTkyaVhYaE1XeTZ6a1JDaEFzdlViTWVyajRpcWo5Z2xKOTBBOU9HNEdxVGw2MGR5STlFdGVXYnRtaVJ6ZTY2OGdqNVNiMkhpUHZrL0VldnlnMlVhV0pKQkZnenE1Q3RXMkI2Qi9WTVUxK2VTdUoyam1hWVB2Um9Obyt0S2cvQUxrM3Y4dWxaMHJmMmdLVlJMLytHUWFYNVoxWElqSk1kRnlnOEpHQ0JscVZWN0FjNHJ1aDBTK1JlenRUODdVSUx1RmtZOERlUEFaYllETld4eit0ZU8vL3RGcFpST0lSYWxKcHJTSFI5WmRab1hRR2JHM0c4dHNmdTlmVFk0Y2FVeDkwbm9xWk1KSTNDb28rWTMxSjVaZnZCN0kxdWIzWGEwVCswZDVwV3BIVnh4eEpyRUdRVEovT3dTRW5WdW05U20vd1hTTCtld0piRERRZGN0ckRoa0NKS212K1gmeHlYTkcvdFJNZmtBNzJDTUQ1SUt3cmJFUFlFPQ=="
	//Csrftoken     = "1d7df6fbd5e80a9b140abd4196ccf0d6"
	//maxRelated    = flag.Int("max_related_pin", 50, "max items per query related")
	conf = flag.String("conf", "./pose_dance_api.toml", "config run file *.toml")
	c    = config.CrawlConfig{}
)

const (
	layout1    = "2006-01-02"
	TimeLayout = "2006-01-12 20:12:00"
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
	err = db.Init(c)
	if err != nil {
		fmt.Println("err when connect postgres", err.Error())
	}
	defer db.Close()
	

	//flag.Parse()
	//configBytes, err := ioutil.ReadFile(*conf)
	//if err != nil {
	//	fmt.Println("err when read config file ", err, "file ", *conf)
	//}
	//err = toml.Unmarshal(configBytes, &c)
	//if err != nil {
	//	fmt.Println("err when pass toml file ", err)
	//}
	//text, err := json.Marshal(c)
	//fmt.Println("Success read config from toml file ", string(text))
	//err = db.Init(c)
	//if err != nil {
	//	fmt.Println("err when connect postgres", err)
	//}
	//defer db.Close()
	//end := time.Now().Unix()
	//println(end)
	////start := end - 24*60*60
	//scores, bestMatchs := db.GetTopUserByMaxScore(10, 10, 0, end)
	//println(scores, len(bestMatchs))
}
