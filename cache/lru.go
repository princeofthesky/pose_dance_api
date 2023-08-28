package cache

import (
	"github.com/hashicorp/golang-lru/v2"
)

var lruUser, _ = lru.New(1000)
var lruMatchResult, _ = lru.New(10000)

func Init(db my)  {

}
func GetUserInfo(user Id)  {

}