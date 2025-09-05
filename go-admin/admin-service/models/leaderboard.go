package models

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// Leaderboard 排行榜配置结构
type Leaderboard struct {
	ID          int64     `orm:"pk;auto" json:"id"`
	AppId       string    `orm:"size(100)" json:"appId"`
	Type        string    `orm:"size(100)" json:"type"`
	Name        string    `orm:"size(200)" json:"name"`
	Description string    `orm:"type(text)" json:"description"`
	ResetType   string    `orm:"size(50)" json:"resetType"` // never, daily, weekly, monthly
	MaxEntries  int       `orm:"default(1000)" json:"maxEntries"`
	ScoreType   string    `orm:"size(20);default(higher_better)" json:"scoreType"` // higher_better, lower_better
	Status      int       `orm:"default(1)" json:"status"`                         // 1=启用, 0=禁用
	CreateTime  time.Time `orm:"auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime  time.Time `orm:"auto_now;type(datetime)" json:"updateTime"`
}

// LeaderboardScore 排行榜分数记录
type LeaderboardScore struct {
	ID            int64     `orm:"pk;auto" json:"id"`
	AppId         string    `orm:"size(100)" json:"appId"`
	LeaderboardId string    `orm:"size(100)" json:"leaderboardId"`
	PlayerId      string    `orm:"size(100)" json:"playerId"`
	OpenId        string    `orm:"size(100)" json:"openId"`
	Score         int64     `json:"score"`
	ExtraData     string    `orm:"type(text)" json:"extraData"`
	HasUserInfo   int       `orm:"default(0)" json:"hasUserInfo"` // 0=无用户信息, 1=有用户信息
	CreateTime    time.Time `orm:"auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime    time.Time `orm:"auto_now;type(datetime)" json:"updateTime"`
}

// GetLeaderboardCount 获取排行榜数量统计
func GetLeaderboardCount(appId string) (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable("leaderboard_config").Filter("app_id", appId).Count()
	return count, err
}

func init() {
	orm.RegisterModel(new(Leaderboard))
	orm.RegisterModel(new(LeaderboardScore))
}

// TableName 获取表名
func (l *Leaderboard) TableName() string {
	return "leaderboard_config"
}

// TableName 获取表名
func (ls *LeaderboardScore) TableName() string {
	return "leaderboard_score"
}

// CreateLeaderboard 创建排行榜
func CreateLeaderboard(leaderboard *Leaderboard) error {
	o := orm.NewOrm()

	// 检查是否已存在
	exist := o.QueryTable("leaderboard_config").
		Filter("app_id", leaderboard.AppId).
		Filter("type", leaderboard.Type).
		Exist()

	if exist {
		return fmt.Errorf("排行榜已存在")
	}

	_, err := o.Insert(leaderboard)
	return err
}

// UpdateLeaderboard 更新排行榜配置
func UpdateLeaderboard(appId, leaderboardType string, fields map[string]interface{}) error {
	o := orm.NewOrm()

	qs := o.QueryTable("leaderboard_config").
		Filter("app_id", appId).
		Filter("type", leaderboardType)

	_, err := qs.Update(fields)
	return err
}

// DeleteLeaderboard 删除排行榜
func DeleteLeaderboard(appId, leaderboardType string) error {
	o := orm.NewOrm()

	// 删除排行榜配置
	_, err := o.QueryTable("leaderboard_config").
		Filter("app_id", appId).
		Filter("type", leaderboardType).
		Delete()

	if err != nil {
		return err
	}

	// 删除排行榜分数记录
	_, err = o.QueryTable("leaderboard_score").
		Filter("app_id", appId).
		Filter("leaderboard_id", leaderboardType).
		Delete()

	return err
}

// GetLeaderboardData 获取排行榜数据
func GetLeaderboardData(appId, leaderboardType string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	o := orm.NewOrm()

	// 获取总数
	total, err := o.QueryTable("leaderboard_score").
		Filter("app_id", appId).
		Filter("leaderboard_id", leaderboardType).
		Count()
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	var scores []LeaderboardScore
	offset := (page - 1) * pageSize
	_, err = o.QueryTable("leaderboard_score").
		Filter("app_id", appId).
		Filter("leaderboard_id", leaderboardType).
		OrderBy("-score", "create_time").
		Limit(pageSize, offset).
		All(&scores)

	if err != nil {
		return nil, 0, err
	}

	// 转换为map格式并添加排名
	var result []map[string]interface{}
	for i, score := range scores {
		item := map[string]interface{}{
			"rank":        offset + i + 1,
			"playerId":    score.PlayerId,
			"openId":      score.OpenId,
			"score":       score.Score,
			"extraData":   score.ExtraData,
			"hasUserInfo": score.HasUserInfo,
			"createTime":  score.CreateTime,
			"updateTime":  score.UpdateTime,
		}
		result = append(result, item)
	}

	return result, total, nil
}

// UpdateLeaderboardScore 更新排行榜分数
func UpdateLeaderboardScore(appId, leaderboardType, playerId string, score int64) error {
	o := orm.NewOrm()

	// 查找现有记录
	var existing LeaderboardScore
	err := o.QueryTable("leaderboard_score").
		Filter("app_id", appId).
		Filter("leaderboard_id", leaderboardType).
		Filter("player_id", playerId).
		One(&existing)

	if err == orm.ErrNoRows {
		// 创建新记录
		newScore := &LeaderboardScore{
			AppId:         appId,
			LeaderboardId: leaderboardType,
			PlayerId:      playerId,
			Score:         score,
		}
		_, err = o.Insert(newScore)
		return err
	} else if err != nil {
		return err
	}

	// 更新现有记录
	existing.Score = score
	existing.UpdateTime = time.Now()
	_, err = o.Update(&existing, "Score", "UpdateTime")
	return err
}

// DeleteLeaderboardScore 删除排行榜分数
func DeleteLeaderboardScore(appId, leaderboardType, playerId string) error {
	o := orm.NewOrm()

	_, err := o.QueryTable("leaderboard_score").
		Filter("app_id", appId).
		Filter("leaderboard_id", leaderboardType).
		Filter("player_id", playerId).
		Delete()

	return err
}

// CommitLeaderboardScore 提交排行榜分数
func CommitLeaderboardScore(appId, leaderboardType, playerId string, score int64) error {
	o := orm.NewOrm()

	// 检查排行榜是否存在且启用
	var leaderboard Leaderboard
	err := o.QueryTable("leaderboard_config").
		Filter("app_id", appId).
		Filter("type", leaderboardType).
		Filter("status", 1).
		One(&leaderboard)

	if err != nil {
		if err == orm.ErrNoRows {
			return fmt.Errorf("排行榜不存在或已禁用")
		}
		return err
	}

	// 查找现有记录
	var existing LeaderboardScore
	err = o.QueryTable("leaderboard_score").
		Filter("app_id", appId).
		Filter("leaderboard_id", leaderboardType).
		Filter("player_id", playerId).
		One(&existing)

	if err == orm.ErrNoRows {
		// 检查是否超过最大条目数
		if leaderboard.MaxEntries > 0 {
			count, _ := o.QueryTable("leaderboard_score").
				Filter("app_id", appId).
				Filter("leaderboard_id", leaderboardType).
				Count()

			if count >= int64(leaderboard.MaxEntries) {
				// 删除最低分记录
				var lowest LeaderboardScore
				err = o.QueryTable("leaderboard_score").
					Filter("app_id", appId).
					Filter("leaderboard_id", leaderboardType).
					OrderBy("score", "-create_time").
					One(&lowest)

				if err == nil && score > lowest.Score {
					o.Delete(&lowest)
				} else if err == nil {
					return fmt.Errorf("分数太低，无法进入排行榜")
				}
			}
		}

		// 创建新记录
		newScore := &LeaderboardScore{
			AppId:         appId,
			LeaderboardId: leaderboardType,
			PlayerId:      playerId,
			Score:         score,
		}
		_, err = o.Insert(newScore)
		return err
	} else if err != nil {
		return err
	}

	// 检查分数类型决定是否更新
	shouldUpdate := false
	if leaderboard.ScoreType == "higher_better" && score > existing.Score {
		shouldUpdate = true
	} else if leaderboard.ScoreType == "lower_better" && score < existing.Score {
		shouldUpdate = true
	}

	if shouldUpdate {
		existing.Score = score
		existing.UpdateTime = time.Now()
		_, err = o.Update(&existing, "Score", "UpdateTime")
	}

	return err
}

// QueryLeaderboardScore 查询排行榜分数
func QueryLeaderboardScore(appId, leaderboardType, playerId string) (int64, int, error) {
	o := orm.NewOrm()

	// 获取用户分数
	var userScore LeaderboardScore
	err := o.QueryTable("leaderboard_score").
		Filter("app_id", appId).
		Filter("leaderboard_id", leaderboardType).
		Filter("player_id", playerId).
		One(&userScore)

	if err != nil {
		if err == orm.ErrNoRows {
			return 0, 0, nil // 用户没有分数记录
		}
		return 0, 0, err
	}

	// 计算排名
	rank, err := o.QueryTable("leaderboard_score").
		Filter("app_id", appId).
		Filter("leaderboard_id", leaderboardType).
		Filter("score__gt", userScore.Score).
		Count()

	if err != nil {
		return userScore.Score, 0, err
	}

	return userScore.Score, int(rank) + 1, nil
}

// FixLeaderboardUserInfo 修复排行榜用户信息
func FixLeaderboardUserInfo(appId, leaderboardType string) (int64, error) {
	o := orm.NewOrm()

	// 获取用户表名
	userTableName := fmt.Sprintf("user_%s", appId)

	// 执行修复SQL
	sql := fmt.Sprintf(`
		UPDATE leaderboard_score ls 
		SET has_user_info = CASE 
			WHEN EXISTS (
				SELECT 1 FROM %s u 
				WHERE u.player_id = ls.player_id 
				AND u.data IS NOT NULL 
				AND u.data != ''
			) THEN 1 
			ELSE 0 
		END 
		WHERE ls.app_id = ? AND ls.leaderboard_id = ?
	`, userTableName)

	result, err := o.Raw(sql, appId, leaderboardType).Exec()
	if err != nil {
		logs.Error("修复排行榜用户信息失败:", err)
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return affected, nil
}
