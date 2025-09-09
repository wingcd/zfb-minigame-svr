package models

import (
	"fmt"
	"time"

	"game-service/utils"

	"github.com/beego/beego/v2/client/orm"
)

// UserDataEntry 用户数据条目模型 - 对应数据库设计的user_data_[appid]表
type UserDataEntry struct {
	ID        int64     `orm:"pk;auto" json:"id"`
	AppId     string    `orm:"-" json:"appId"`                                             // 应用ID（仅用于逻辑，不存储到数据库）
	UserID    string    `orm:"size(100);column(user_id)" json:"userID"`                    // 用户ID
	DataKey   string    `orm:"size(100);column(data_key)" json:"dataKey"`                  // 数据键
	DataValue string    `orm:"type(longtext);column(data_value)" json:"dataValue"`         // 数据值（JSON格式）
	DataType  string    `orm:"size(50);default(string);column(data_type)" json:"dataType"` // 数据类型
	IsPublic  bool      `orm:"default(false);column(is_public)" json:"isPublic"`           // 是否公开
	Tags      string    `orm:"type(text)" json:"tags"`                                     // 标签（JSON数组）
	Version   int       `orm:"default(1)" json:"version"`                                  // 版本号
	CreatedAt time.Time `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

// GetTableName 获取动态表名
func (ude *UserDataEntry) GetTableName(appId string) string {
	return utils.GetUserDataTableName(appId)
}

// SaveUserDataWithKey 保存用户数据（指定key）
func SaveUserDataWithKey(appId, userId, dataKey, dataValue string) error {
	o := orm.NewOrm()
	tableName := utils.GetUserDataTableName(appId)

	// 使用 ON DUPLICATE KEY UPDATE 进行 upsert 操作
	sql := fmt.Sprintf(`
		INSERT INTO %s (user_id, data_key, data_value, data_type, version, created_at, updated_at)
		VALUES (?, ?, ?, 'string', 1, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			data_value = VALUES(data_value),
			version = version + 1,
			updated_at = NOW()
	`, tableName)

	_, err := o.Raw(sql, userId, dataKey, dataValue).Exec()
	return err
}

// GetUserDataWithKey 获取用户数据（指定key）
func GetUserDataWithKey(appId, userId, dataKey string) (string, error) {
	o := orm.NewOrm()
	tableName := utils.GetUserDataTableName(appId)

	sql := fmt.Sprintf(`
		SELECT data_value FROM %s 
		WHERE user_id = ? AND data_key = ?
	`, tableName)

	var dataValue string
	err := o.Raw(sql, userId, dataKey).QueryRow(&dataValue)
	if err == orm.ErrNoRows {
		return "", nil // 返回空字符串表示无数据
	} else if err != nil {
		return "", err
	}

	return dataValue, nil
}

// DeleteUserDataWithKey 删除用户数据（指定key）
func DeleteUserDataWithKey(appId, userId, dataKey string) error {
	o := orm.NewOrm()
	tableName := utils.GetUserDataTableName(appId)

	sql := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE user_id = ? AND data_key = ?
	`, tableName)

	_, err := o.Raw(sql, userId, dataKey).Exec()
	return err
}

// DeleteAllUserData 删除用户的所有数据
func DeleteAllUserData(appId, userId string) error {
	o := orm.NewOrm()
	tableName := utils.GetUserDataTableName(appId)

	sql := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE user_id = ?
	`, tableName)

	_, err := o.Raw(sql, userId).Exec()
	return err
}

// GetUserDataList 获取用户数据列表（管理后台使用）
func GetUserDataList(appId string, page, pageSize int) ([]UserDataEntry, int64, error) {
	o := orm.NewOrm()
	tableName := utils.GetUserDataTableName(appId)

	// 查询总数
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, tableName)

	var total int64
	err := o.Raw(countSQL).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	offset := (page - 1) * pageSize
	dataSQL := fmt.Sprintf(`
		SELECT * FROM %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, tableName)

	var results []orm.Params
	_, err = o.Raw(dataSQL, pageSize, offset).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 转换为UserDataEntry结构
	var dataList []UserDataEntry
	for _, result := range results {
		entry := UserDataEntry{AppId: appId}

		if id, ok := result["id"].(int64); ok {
			entry.ID = id
		}
		if userId, ok := result["user_id"].(string); ok {
			entry.UserID = userId
		}
		if dataKey, ok := result["data_key"].(string); ok {
			entry.DataKey = dataKey
		}
		if dataValue, ok := result["data_value"].(string); ok {
			entry.DataValue = dataValue
		}
		if dataType, ok := result["data_type"].(string); ok {
			entry.DataType = dataType
		}
		if isPublic, ok := result["is_public"].(bool); ok {
			entry.IsPublic = isPublic
		}
		if tags, ok := result["tags"].(string); ok {
			entry.Tags = tags
		}
		if version, ok := result["version"].(int64); ok {
			entry.Version = int(version)
		}
		if createdAt, ok := result["created_at"].(time.Time); ok {
			entry.CreatedAt = createdAt
		}
		if updatedAt, ok := result["updated_at"].(time.Time); ok {
			entry.UpdatedAt = updatedAt
		}

		dataList = append(dataList, entry)
	}

	return dataList, total, nil
}

// GetAllUserDataEntries 获取用户的所有数据条目
func GetAllUserDataEntries(appId, userId string) ([]UserDataEntry, error) {
	o := orm.NewOrm()
	tableName := utils.GetUserDataTableName(appId)

	sql := fmt.Sprintf(`
		SELECT * FROM %s 
		WHERE user_id = ?
		ORDER BY data_key, updated_at DESC
	`, tableName)

	var results []orm.Params
	_, err := o.Raw(sql, userId).Values(&results)
	if err != nil {
		return nil, err
	}

	var entries []UserDataEntry
	for _, result := range results {
		entry := UserDataEntry{AppId: appId}

		if id, ok := result["id"].(int64); ok {
			entry.ID = id
		}
		if userId, ok := result["user_id"].(string); ok {
			entry.UserID = userId
		}
		if dataKey, ok := result["data_key"].(string); ok {
			entry.DataKey = dataKey
		}
		if dataValue, ok := result["data_value"].(string); ok {
			entry.DataValue = dataValue
		}
		if dataType, ok := result["data_type"].(string); ok {
			entry.DataType = dataType
		}
		if isPublic, ok := result["is_public"].(bool); ok {
			entry.IsPublic = isPublic
		}
		if tags, ok := result["tags"].(string); ok {
			entry.Tags = tags
		}
		if version, ok := result["version"].(int64); ok {
			entry.Version = int(version)
		}
		if createdAt, ok := result["created_at"].(time.Time); ok {
			entry.CreatedAt = createdAt
		}
		if updatedAt, ok := result["updated_at"].(time.Time); ok {
			entry.UpdatedAt = updatedAt
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
