package db

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go-pinterest/config"
	"strconv"
)

var rdb *redis.Client
var redisEmpty = "redis: nil"

func initRedis(config config.Redis) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password, // no password set
		DB:       config.Db,       // use default DB
	})
	return rdb.Ping(context.Background()).Err()
}

func closeRedis() error {
	return rdb.Close()
}

//func UpdateTotalScoreAvgByUser(userId int, audioId int, score int) (float64, int, error) {
//	totalKey := "total_" + strconv.Itoa(audioId)
//	countKey := "count_" + strconv.Itoa(audioId)
//	allStats, err := rdb.HMGet(context.Background(), hKeyUsersStatInfo(userId), totalKey, countKey).Result()
//	if err != nil && err.Error() != redisEmpty {
//		println("HMGEt error : ", err.Error())
//		return 0, 0, err
//	}
//	totalScoreOldValue, err := rdb.ZScore(context.Background(), zKeyUsersStat(), strconv.Itoa(userId)).Result()
//	if err != nil {
//		if err.Error() != redisEmpty {
//			println("ZScore error : ", err.Error())
//			return 0, 0, err
//		} else {
//			totalScoreOldValue = 0
//		}
//	}
//	if totalScoreOldValue <= 0 {
//		totalScoreOldValue = 0
//	}
//	totalScoreAudioOldValue := int64(0)
//	totalCountAudioOldValue := int64(0)
//	if allStats[0] != nil {
//		totalScoreAudioOldValueText := allStats[0].(string)
//		totalScoreAudioOldValue, _ = strconv.ParseInt(totalScoreAudioOldValueText, 10, 64)
//	}
//	if allStats[1] != nil {
//		totalCountAudioOldValueText := allStats[1].(string)
//		totalCountAudioOldValue, _ = strconv.ParseInt(totalCountAudioOldValueText, 10, 64)
//	}
//	totalScoreAudioNewValue := totalScoreAudioOldValue + int64(score)
//	totalCountAudioNewValue := totalCountAudioOldValue + 1
//
//	totalScoreNewValue := float64(0)
//	if totalCountAudioOldValue == 0 {
//		totalScoreNewValue = totalScoreOldValue + float64(totalScoreAudioNewValue)/float64(totalCountAudioNewValue)
//	} else {
//		totalScoreNewValue = totalScoreOldValue -
//			float64(totalScoreAudioOldValue)/float64(totalCountAudioOldValue) +
//			float64(totalScoreAudioNewValue)/float64(totalCountAudioNewValue)
//	}
//
//	if totalScoreOldValue <= 0 {
//		totalScoreOldValue = 0
//	}
//
//	pipe := rdb.Pipeline()
//	pipe.HMSet(context.Background(), hKeyUsersStatInfo(userId), totalKey, totalScoreAudioNewValue, countKey, totalCountAudioNewValue)
//	pipe.ZAdd(context.Background(), zKeyUsersStat(), redis.Z{Member: userId, Score: totalScoreNewValue})
//	_, err = pipe.Exec(context.Background())
//	if err != nil {
//		println("pipe Exec error : ", err.Error())
//	}
//	return GetRankByUser(userId)
//}

func UpdateTotalScoreByUser(userId int, score int) (float64, int, error) {
	totalScore, err := rdb.ZIncrBy(context.Background(), zKeyUsersStat(), float64(score), strconv.Itoa(userId)).Result()
	if err != nil {
		println("ZIncrBy error : ", err.Error())
		return 0, -1, err
	}
	rank, err := rdb.ZRevRank(context.Background(), zKeyUsersStat(), strconv.Itoa(userId)).Result()
	if err != nil {
		println("ZScore error : ", err.Error())
		return 0, -1, err
	}
	return totalScore, int(rank) + 1, nil
}

func GetRankByUser(userId int) (float64, int, error) {
	rank, err := rdb.ZRevRank(context.Background(), zKeyUsersStat(), strconv.Itoa(userId)).Result()
	if err != nil {
		println("ZScore error : ", err.Error())
		return 0, -1, err
	}
	score, err := rdb.ZScore(context.Background(), zKeyUsersStat(), strconv.Itoa(userId)).Result()
	if err != nil {
		println("ZScore error : ", err.Error())
		return 0, -1, err
	}
	return score, int(rank) + 1, nil
}

func RemoveUserInfoInRedis(userId int) {
	//_, err := rdb.HDel(context.Background(), hKeyUsersStatInfo(userId)).Result()
	//if err != nil {
	//	println("Error when remove stats in redis by user id ", userId, " err : ", err.Error())
	//}
	_, err := rdb.ZRem(context.Background(), zKeyUsersStat(), strconv.Itoa(userId)).Result()
	if err != nil {
		println("Error when remove rank in redis  by user id ", userId, " err : ", err.Error())
	}
}

func MergeUserInfoInRedis(oldId int, newId int) float64 {
	totalOldScore, err := rdb.ZScore(context.Background(), zKeyUsersStat(), strconv.Itoa(oldId)).Result()
	if err != nil {
		println("ZScore error : ", err.Error())
	}
	pipe := rdb.Pipeline()
	pipe.ZIncrBy(context.Background(), zKeyUsersStat(), float64(totalOldScore), strconv.Itoa(newId))
	pipe.ZRem(context.Background(), zKeyUsersStat(), strconv.Itoa(oldId)).Result()
	_, err = pipe.Exec(context.Background())
	if err != nil {
		println("Error when merge user rank in redis  by old id ", oldId, "new_id", newId, " err : ", err.Error())
	}
	return totalOldScore
}
