package db

import (
	"context"
	"github.com/go-pg/pg/v10"
	"go-pinterest/config"
	"sort"
)

var mysqlDb *pg.DB

func GetMysqlDb() *pg.DB {
	return mysqlDb
}
func initMysql(postgres config.Postgres) error {
	mysqlDb = pg.Connect(&pg.Options{
		Addr:     postgres.Addr,
		User:     postgres.User,
		Password: postgres.Password,
		Database: postgres.Database,
	})
	return mysqlDb.Ping(context.Background())
}

func closeMysql() {
	mysqlDb.Close()
}

func GetDb() *pg.DB {
	return mysqlDb
}

func GetUserById(id int) (User, error) {
	info := User{Id: id}
	err := mysqlDb.Model(&info).WherePK().Select()
	return info, err
}

func GetUserInfoByUserName(username string, provider string) (User, error) {
	info := User{UserName: username, Provider: provider}
	err := mysqlDb.Model(&info).Where("user_name= ? AND provider= ?", info.UserName, info.Provider).Select()
	return info, err
}

func InsertUserInfo(info User) (User, error) {
	_, err := mysqlDb.Model(&info).Insert()
	return info, err
}

func GetUserAndStatById(id int) (UserAndStat, error) {
	info := UserAndStat{UserId: id}
	err := mysqlDb.Model(&info).Where("user_id = ? ", info.UserId).Select()
	return info, err
}

func InsertUserAndStatInfo(info UserAndStat) (UserAndStat, error) {
	_, err := mysqlDb.Model(&info).OnConflict("(user_id) DO UPDATE").
		Set("total_game = ? , time_avg = ? , accuracy_avg = ? ", info.TotalGame, info.TimeAvg, info.AccuracyAvg).Insert()
	return info, err
}

func UpdateAvatarUserInfo(info User) (User, error) {
	_, err := mysqlDb.Model(&info).WherePK().Column("avatar").Update()
	return info, err
}

func UpdateNameUserInfo(info User) (User, error) {
	_, err := mysqlDb.Model(&info).WherePK().Column("name").Update()
	return info, err
}

func GetMatchResultById(id int) (MatchResult, error) {
	info := MatchResult{Id: id}
	err := mysqlDb.Model(&info).WherePK().Select()
	return info, err
}

func GetMatchResultByVideoMd5(md5 string) (MatchResult, error) {
	info := MatchResult{Id: 0, VideoMd5: md5}
	err := mysqlDb.Model(&info).Where("video_md5=?", info.VideoMd5).Select()
	return info, err
}

func InsertMatchResult(info MatchResult) (MatchResult, error) {
	_, err := mysqlDb.Model(&info).Insert()
	return info, err
}

func UpdateVideoMatchResult(info MatchResult) (MatchResult, error) {
	_, err := mysqlDb.Model(&info).WherePK().Column("video", "cover", "video_md5", "cover_md5").Update()
	return info, err
}

func GetListMatchResultId(offset int, length int) ([]MatchResult, error) {
	data := []MatchResult{}
	err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("user_id > 0  AND created_time < '?'", offset).Limit(length).Order("created_time DESC").ForEach(
		func(c *MatchResult) error {
			data = append(data, *c)
			return nil
		})
	return data, err
}

func GetTopMatchResultByAudioId(audioId string, playMode string,score int, createdTime int64, length int) ([]MatchResult, error) {
	data := []MatchResult{}
	if len(playMode) ==0 {
		err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("user_id > 0 AND audio_id = ? AND score<= ? AND received_time > ?", audioId, score, createdTime).Order("score DESC").Order("received_time ASC").Limit(length).ForEach(
			func(c *MatchResult) error {
				data = append(data, *c)
				return nil
			})
		return data, err
	}else {
		err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("user_id > 0 AND audio_id = ? AND score<= ? AND received_time > ? AND play_mode = ? ", audioId, score, createdTime,playMode).Order("score DESC").Order("received_time ASC").Limit(length).ForEach(
			func(c *MatchResult) error {
				data = append(data, *c)
				return nil
			})
		return data, err
	}
}

func GetTopVideoMatchResultByAudioId(audioId string,playMode string, score int, createdTime int64, length int) ([]MatchResult, error) {
	data := []MatchResult{}
	if len(playMode) ==0 {
		err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("user_id > 0 AND audio_id = ? AND score<= ? AND received_time > ? AND length(video)>0", audioId, score, createdTime).Order("score DESC").Order("received_time ASC").Limit(length).ForEach(
			func(c *MatchResult) error {
				data = append(data, *c)
				return nil
			})
		return data, err
	}else {
		err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("user_id > 0 AND audio_id = ? AND score<= ? AND received_time > ? AND length(video)>0 AND play_mode = ?", audioId, score, createdTime,playMode).Order("score DESC").Order("received_time ASC").Limit(length).ForEach(
			func(c *MatchResult) error {
				data = append(data, *c)
				return nil
			})
		return data, err
	}
}

func GetTopVideoMatchResultByAudioIdAndLimitScore(audioId int,playMode string, lowerScore int, sortType string, offset int) ([]MatchResult, error) {
	data := []MatchResult{}
	if audioId > 0 {
		if len(playMode) ==0 {
			err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("score > ? AND audio_id = ?  AND length(video)>0", lowerScore, audioId).Order(sortType + " DESC").Limit(20).Offset(offset).ForEach(
				func(c *MatchResult) error {
					data = append(data, *c)
					return nil
				})
			return data, err
		}else {
			err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("score > ? AND audio_id = ?  AND length(video)>0 AND play_mode = ? ", lowerScore, audioId,playMode).Order(sortType + " DESC").Limit(20).Offset(offset).ForEach(
				func(c *MatchResult) error {
					data = append(data, *c)
					return nil
				})
			return data, err
		}
	} else {
		if len(playMode) == 0 {
			err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("score > ? AND length(video)>0", lowerScore).Order(sortType + " DESC").Limit(20).Offset(offset).ForEach(
				func(c *MatchResult) error {
					data = append(data, *c)
					return nil
				})
			return data, err
		}else {
			err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("score > ? AND length(video)>0  AND play_mode = ? ", lowerScore,playMode).Order(sortType + " DESC").Limit(20).Offset(offset).ForEach(
				func(c *MatchResult) error {
					data = append(data, *c)
					return nil
				})
			return data, err
		}
	}
}

func GetTopVideoMatchResultByAudioIdAndLimitScoreInATime(audioId int, lowerScore int, sortType string, startTime int, endTime int64, offset int, limit int) ([]MatchResult, error) {
	data := []MatchResult{}
	if audioId > 0 {
		err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("score > ? AND audio_id = ?  AND length(video)>0 AND received_time >= ? AND received_time < ? ", lowerScore, audioId, startTime, endTime).Order(sortType + " DESC").Limit(limit).Offset(offset).ForEach(
			func(c *MatchResult) error {
				data = append(data, *c)
				return nil
			})
		return data, err
	} else {
		err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("score > ? AND length(video)>0 AND received_time >= ? AND received_time < ? ", lowerScore, startTime, endTime).Order(sortType + " DESC").Limit(limit).Offset(offset).ForEach(
			func(c *MatchResult) error {
				data = append(data, *c)
				return nil
			})
		return data, err
	}
}

func GetNewstVideoMatchResultByAudioIdAndLimitScoreInATime(audioId int, lowerScore int, limitTime int, offset int, limit int) ([]MatchResult, error) {
	data := []MatchResult{}
	err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("(score is not NULL || score > ?) AND audio_id = ?  AND length(video)>0 AND received_time >= ? ", lowerScore, audioId, limitTime).Order("received_time DESC").Limit(limit).Offset(offset).ForEach(
		func(c *MatchResult) error {
			data = append(data, *c)
			return nil
		})
	return data, err

}

func GetTotalMatchResultByAudioId(audioId string) (int, error) {
	result, err := mysqlDb.Model((*MatchResult)(nil)).Where("audio_id = ? ", audioId).Count()
	return result, err
}

func GetTotalPlayerByAudioId(audioId string) (int, error) {
	result, err := mysqlDb.Model((*MatchResult)(nil)).Where("audio_id = ? ", audioId).DistinctOn("user_id").Count()
	return result, err
}

func GetBestMatchResultByAudioId(audioId string, limit int) ([]MatchResult, error) {
	var result []MatchResult
	err := mysqlDb.Model((*MatchResult)(nil)).Where("audio_id = ? AND score is NOT NULL", audioId).Order("score DESC").Limit(limit).ForEach(
		func(c *MatchResult) error {
			result = append(result, *c)
			return nil
		})
	return result, err
}
func GetNewVideoMatchResultByAudioId(audioId string, createdTime int64, length int) ([]MatchResult, error) {
	data := []MatchResult{}
	err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("audio_id = ? AND received_time < ? AND length(video)>0", audioId, createdTime).Order("received_time DESC").Limit(length).ForEach(
		func(c *MatchResult) error {
			data = append(data, *c)
			return nil
		})
	return data, err
}

func GetTopMatchResultByUser(userId int, score int, createdTime int64, length int) ([]MatchResult, error) {
	data := []MatchResult{}
	err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("user_id = ? AND score<= ? AND received_time > ?", userId, score, createdTime).Order("score DESC").Order("received_time ASC").Limit(length).ForEach(
		func(c *MatchResult) error {
			data = append(data, *c)
			return nil
		})
	return data, err
}

func GetTopVideoMatchResultByUser(userId int, score int, createdTime int64, length int) ([]MatchResult, error) {
	data := []MatchResult{}
	err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("user_id = ? AND score<= ? AND received_time > ? AND length(video)>0 ", userId, score, createdTime).Order("score DESC").Order("received_time ASC").Limit(length).ForEach(
		func(c *MatchResult) error {
			data = append(data, *c)
			return nil
		})
	return data, err
}

func GetNewVideoMatchResultByUser(userId int, createdTime int64, length int) ([]MatchResult, error) {
	data := []MatchResult{}
	err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("user_id = ? AND received_time < ? AND length(video)>0 ", userId, createdTime).Order("received_time DESC").Limit(length).ForEach(
		func(c *MatchResult) error {
			data = append(data, *c)
			return nil
		})
	return data, err
}

func RemoveUserInfoInMysql(userId int) {
	_, err := mysqlDb.Exec("delete from match_results where user_id = ? ", userId)
	if err != nil {
		println("Error when remove match result by user id ", userId, " err : ", err.Error())
	}
	_, err = mysqlDb.Exec("delete from user_and_stats where user_id = ? ", userId)
	if err != nil {
		println("Error when remove stats  by user id ", userId, " err : ", err.Error())
	}
	_, err = mysqlDb.Exec("delete from users where id = ? ", userId)
	if err != nil {
		println("Error when remove user info by user id ", userId, " err : ", err.Error())
	}
}

func RemoveMatchInfoInMysql(matchId int) {
	_, err := mysqlDb.Exec("delete from match_results where id = ? ", matchId)
	if err != nil {
		println("Error when remove match result by id ", matchId, " err : ", err.Error())
	}
}


func MergerUserInfoInMysql(oldId int, newId int) {
	_, err := mysqlDb.Exec("update match_results set user_id = ? where user_id = ? ", newId, oldId)
	if err != nil {
		println("Error update match result by old_id ", oldId, " new_id ", newId, " err : ", err.Error())
	}
	oldStat, _ := GetUserAndStatById(oldId)
	newStat, _ := GetUserAndStatById(newId)
	newStat.TimeAvg = (float32(oldStat.TotalGame)*oldStat.TimeAvg + float32(newStat.TotalGame)*newStat.TimeAvg) / float32(newStat.TotalGame+oldStat.TotalGame)
	newStat.AccuracyAvg = (float32(oldStat.TotalGame)*oldStat.AccuracyAvg + float32(newStat.TotalGame)*newStat.AccuracyAvg) / float32(newStat.TotalGame+oldStat.TotalGame)
	newStat.TotalGame = newStat.TotalGame + oldStat.TotalGame

	oldStat.TotalGame = 0
	oldStat.TimeAvg = 0
	oldStat.AccuracyAvg = 0
	_, err = InsertUserAndStatInfo(newStat)
	if err != nil {
		println("Error update new user stat when merge user  old_id ", oldId, " new_id ", newId, " err : ", err.Error())
	}
	_, err = InsertUserAndStatInfo(oldStat)
	if err != nil {
		println("Error update old user stat when merge user  old_id ", oldId, " new_id ", newId, " err : ", err.Error())
	}
}

func GetTopVideoMatchResultByTime(offset int64) ([]MatchResult, error) {
	data := []MatchResult{}
	err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("received_time < ? AND length(video)>0", offset).Order("received_time DESC").Limit(20).ForEach(
		func(c *MatchResult) error {
			data = append(data, *c)
			return nil
		})
	return data, err
}

func GetUserIdByVideoUser(videoFile string) (int, error) {
	userId := -2
	err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("strpos (video,?) > 0", videoFile).Limit(1).ForEach(
		func(c *MatchResult) error {
			userId = c.UserId
			return nil
		})
	return userId, err
}

func GetMatchResultByVideoFileOrCoverFile(file string) (MatchResult, error) {
	result := MatchResult{}
	err := mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("strpos (video,?) > 0 OR strpos (cover,?) > 0", file, file).Limit(1).ForEach(
		func(c *MatchResult) error {
			result = *c
			return nil
		})
	return result, err
}

func GetTopUserByMaxScore(limitUser int, limitMatch int, startTime int64, endTime int64) ([]int, []MatchResult) {
	mapScoreUsers := map[int][]int{}
	mapBestMatches := map[int]MatchResult{}
	mysqlDb.Model((*MatchResult)(nil)).Column("*").Where("received_time >=? AND  received_time <= ? AND length(video)>0 ", startTime, endTime).ForEach(
		func(c *MatchResult) error {
			scores := mapScoreUsers[c.UserId]
			scores = append(scores, c.Score)
			mapScoreUsers[c.UserId] = scores
			bestMatch, exit := mapBestMatches[c.UserId]
			if !exit || bestMatch.Score < c.Score {
				mapBestMatches[c.UserId] = *c
			}
			return nil
		})
	type scoreUser struct {
		userId int
		total  int
	}
	rankUser := []scoreUser{}
	for userId, scoresUser := range mapScoreUsers {
		sort.Slice(scoresUser, func(i, j int) bool {
			return scoresUser[i] > scoresUser[j]
		})
		total := 0
		for i := 0; i < limitMatch; i++ {
			if i < len(scoresUser) {
				total = total + scoresUser[i]
			}
		}
		rankUser = append(rankUser, scoreUser{total: total, userId: userId})
	}
	sort.Slice(rankUser, func(i, j int) bool {
		return rankUser[i].total > rankUser[j].total
	})
	scores := []int{}
	bestMatches := []MatchResult{}
	for i := 0; i < limitUser; i++ {
		if i < len(rankUser) {
			userId := rankUser[i].userId
			println(rankUser[i].userId, rankUser[i].total, mapBestMatches[userId].UserId)
			scores = append(scores, rankUser[i].total)
			bestMatches = append(bestMatches, mapBestMatches[rankUser[i].userId])
		}
	}
	return scores, bestMatches
}

func GetAllYoutubeMatch() ([]MatchAndYoutube,error){
	var result []MatchAndYoutube
	err := mysqlDb.Model((*MatchAndYoutube)(nil)).Column("*").ForEach(
		func(c *MatchAndYoutube) error {
			result =append(result,*c)
			return nil
		})
	return result, err
}