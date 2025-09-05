package models

import (
	"fmt"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

// Leaderboard 排行榜模型
type Leaderboard struct {
	Id              int64  `orm:"auto" json:"id"`
	LeaderboardName string `orm:"size(100)" json:"leaderboard_name"`
	UserId          string `orm:"size(100)" json:"user_id"`
	Score           int64  `orm:"default(0)" json:"score"`
	ExtraData       string `orm:"type(text)" json:"extra_data"`
	CreatedAt       string `orm:"auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt       string `orm:"auto_now;type(datetime)" json:"updated_at"`
}

// GetTableName 获取动态表名
func (l *Leaderboard) GetTableName(appId string) string {
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	return fmt.Sprintf("leaderboard_%s", cleanAppId)
}

// SubmitScore 提交分数到排行榜
func SubmitScore(appId, userId, leaderboardName string, score int64, extraData string) error {
	o := orm.NewOrm()

	leaderboard := &Leaderboard{}
	tableName := leaderboard.GetTableName(appId)

	// 检查是否已存在记录
	err := o.QueryTable(tableName).
		Filter("leaderboard_name", leaderboardName).
		Filter("user_id", userId).
		One(leaderboard)

	switch err {
	case orm.ErrNoRows:
		// 新建记录
		leaderboard.LeaderboardName = leaderboardName
		leaderboard.UserId = userId
		leaderboard.Score = score
		leaderboard.ExtraData = extraData
		_, err = o.Insert(leaderboard)
	case nil:
		leaderboard.Score = score
		leaderboard.ExtraData = extraData
		_, err = o.Update(leaderboard, "score", "extra_data", "updated_at")
	}

	return err
}

func UpdateScore(appId, userId, leaderboardName string, score int64, extraData string) error {
	o := orm.NewOrm()

	leaderboard := &Leaderboard{}
	tableName := leaderboard.GetTableName(appId)

	err := o.QueryTable(tableName).
		Filter("leaderboard_name", leaderboardName).
		Filter("user_id", userId).
		One(leaderboard)

	switch err {
	case orm.ErrNoRows:
		return fmt.Errorf("排行榜不存在")
	case nil:
		leaderboard.Score = score
		leaderboard.ExtraData = extraData
		_, err = o.Update(leaderboard, "score", "extra_data", "updated_at")
	}

	return err
}

// GetLeaderboard 获取排行榜
func GetLeaderboard(appId, leaderboardName string, limit int) ([]Leaderboard, error) {
	o := orm.NewOrm()

	leaderboard := &Leaderboard{}
	tableName := leaderboard.GetTableName(appId)

	var results []Leaderboard
	_, err := o.QueryTable(tableName).
		Filter("leaderboard_name", leaderboardName).
		OrderBy("-score", "created_at").
		Limit(limit).
		All(&results)

	return results, err
}

// GetUserRank 获取用户在排行榜中的排名
func GetUserRank(appId, userId, leaderboardName string) (int, int64, error) {
	o := orm.NewOrm()

	leaderboard := &Leaderboard{}
	tableName := leaderboard.GetTableName(appId)

	// 获取用户分数
	var userScore int64
	err := o.QueryTable(tableName).
		Filter("leaderboard_name", leaderboardName).
		Filter("user_id", userId).
		One(leaderboard)

	if err == orm.ErrNoRows {
		return 0, 0, nil // 用户不在排行榜中
	} else if err != nil {
		return 0, 0, err
	}

	userScore = leaderboard.Score

	// 计算排名
	var rank int64
	err = o.Raw(fmt.Sprintf(`
		SELECT COUNT(*) + 1 FROM %s 
		WHERE leaderboard_name = ? AND score > ?
	`, tableName), leaderboardName, userScore).QueryRow(&rank)

	if err != nil {
		return 0, 0, err
	}

	return int(rank), userScore, nil
}

// ResetLeaderboard 重置排行榜
func ResetLeaderboard(appId, leaderboardName string) error {
	o := orm.NewOrm()

	leaderboard := &Leaderboard{}
	tableName := leaderboard.GetTableName(appId)

	_, err := o.QueryTable(tableName).Filter("leaderboard_name", leaderboardName).Delete()
	return err
}

// GetLeaderboardList 获取排行榜列表（管理后台使用）
func GetLeaderboardList(appId string, page, pageSize int, leaderboardName string) ([]Leaderboard, int64, error) {
	o := orm.NewOrm()

	leaderboard := &Leaderboard{}
	tableName := leaderboard.GetTableName(appId)

	qs := o.QueryTable(tableName)
	if leaderboardName != "" {
		qs = qs.Filter("leaderboard_name", leaderboardName)
	}

	total, _ := qs.Count()

	var results []Leaderboard
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-score", "created_at").Limit(pageSize, offset).All(&results)

	return results, total, err
}
