package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"game-service/models"
	"game-service/utils"

	"github.com/beego/beego/v2/client/orm"
)

// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{}
}

// SaveUserData 保存用户数据
func (s *UserService) SaveUserData(appId, userId string, data interface{}) error {
	o := orm.NewOrm()

	// 将数据转换为JSON字符串
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("数据序列化失败: %v", err)
	}
	dataStr := string(dataBytes)

	// 获取表名
	tableName := s.getUserDataTableName(appId)

	// 检查用户是否已存在
	var existingId int64
	checkSQL := fmt.Sprintf("SELECT id FROM %s WHERE user_id = ?", tableName)
	err = o.Raw(checkSQL, userId).QueryRow(&existingId)

	if err == sql.ErrNoRows {
		// 用户不存在，插入新记录
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s (user_id, data, create_time, update_time) 
			VALUES (?, ?, NOW(), NOW())
		`, tableName)
		_, err = o.Raw(insertSQL, userId, dataStr).Exec()
		if err != nil {
			return fmt.Errorf("保存用户数据失败: %v", err)
		}
	} else if err == nil {
		// 用户已存在，更新记录
		updateSQL := fmt.Sprintf(`
			UPDATE %s SET data = ?, update_time = NOW() WHERE user_id = ?
		`, tableName)
		_, err = o.Raw(updateSQL, dataStr, userId).Exec()
		if err != nil {
			return fmt.Errorf("更新用户数据失败: %v", err)
		}
	} else {
		return fmt.Errorf("查询用户数据失败: %v", err)
	}

	return nil
}

// GetUserData 获取用户数据
func (s *UserService) GetUserData(appId, userId string) (interface{}, error) {
	o := orm.NewOrm()

	tableName := s.getUserDataTableName(appId)

	var dataStr string
	querySQL := fmt.Sprintf("SELECT data FROM %s WHERE user_id = ?", tableName)
	err := o.Raw(querySQL, userId).QueryRow(&dataStr)

	if err == sql.ErrNoRows {
		// 用户不存在，返回空数据
		return map[string]interface{}{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("查询用户数据失败: %v", err)
	}

	// 解析JSON数据
	var data interface{}
	if dataStr != "" {
		err = json.Unmarshal([]byte(dataStr), &data)
		if err != nil {
			return nil, fmt.Errorf("数据反序列化失败: %v", err)
		}
	} else {
		data = map[string]interface{}{}
	}

	return data, nil
}

// DeleteUserData 删除用户数据
func (s *UserService) DeleteUserData(appId, userId string) error {
	o := orm.NewOrm()

	tableName := s.getUserDataTableName(appId)

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", tableName)
	result, err := o.Raw(deleteSQL, userId).Exec()
	if err != nil {
		return fmt.Errorf("删除用户数据失败: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// GetUserList 获取用户列表（管理后台使用）
func (s *UserService) GetUserList(appId string, page, pageSize int, keyword string) (*models.PageData, error) {
	o := orm.NewOrm()

	tableName := s.getUserDataTableName(appId)

	// 构建查询条件
	whereClause := ""
	args := []interface{}{}
	if keyword != "" {
		whereClause = "WHERE user_id LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	// 统计总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tableName, whereClause)
	var total int64
	err := o.Raw(countSQL, args...).QueryRow(&total)
	if err != nil {
		return nil, fmt.Errorf("统计用户数量失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	querySQL := fmt.Sprintf(`
		SELECT id, user_id, create_time, update_time 
		FROM %s %s 
		ORDER BY id DESC 
		LIMIT ? OFFSET ?
	`, tableName, whereClause)
	args = append(args, pageSize, offset)

	var users []orm.Params
	_, err = o.Raw(querySQL, args...).Values(&users)
	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %v", err)
	}

	return &models.PageData{
		List:     users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetUserDetail 获取用户详细信息（管理后台使用）
func (s *UserService) GetUserDetail(appId, userId string) (map[string]interface{}, error) {
	o := orm.NewOrm()

	tableName := s.getUserDataTableName(appId)

	var result []orm.Params
	querySQL := fmt.Sprintf("SELECT * FROM %s WHERE user_id = ?", tableName)
	_, err := o.Raw(querySQL, userId).Values(&result)
	if err != nil {
		return nil, fmt.Errorf("查询用户详情失败: %v", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("用户不存在")
	}

	userDetail := make(map[string]interface{})
	for k, v := range result[0] {
		userDetail[k] = v
	}

	// 解析用户数据JSON
	if dataStr, ok := userDetail["data"].(string); ok && dataStr != "" {
		var userData interface{}
		if err := json.Unmarshal([]byte(dataStr), &userData); err == nil {
			userDetail["parsed_data"] = userData
		}
	}

	return userDetail, nil
}

// BatchDeleteUsers 批量删除用户（管理后台使用）
func (s *UserService) BatchDeleteUsers(appId string, userIds []string) error {
	if len(userIds) == 0 {
		return fmt.Errorf("用户ID列表不能为空")
	}

	o := orm.NewOrm()
	tableName := s.getUserDataTableName(appId)

	// 构建IN条件
	placeholders := utils.BuildPlaceholders(len(userIds))
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE user_id IN (%s)", tableName, placeholders)

	// 转换参数类型
	args := make([]interface{}, len(userIds))
	for i, id := range userIds {
		args[i] = id
	}

	_, err := o.Raw(deleteSQL, args...).Exec()
	if err != nil {
		return fmt.Errorf("批量删除用户失败: %v", err)
	}

	return nil
}

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats(appId string) (map[string]interface{}, error) {
	o := orm.NewOrm()

	tableName := s.getUserDataTableName(appId)
	stats := make(map[string]interface{})

	// 总用户数
	var totalUsers int64
	totalSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := o.Raw(totalSQL).QueryRow(&totalUsers)
	if err == nil {
		stats["total_users"] = totalUsers
	}

	// 今日新增用户数
	var todayUsers int64
	todaySQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE DATE(create_time) = CURDATE()", tableName)
	err = o.Raw(todaySQL).QueryRow(&todayUsers)
	if err == nil {
		stats["today_new_users"] = todayUsers
	}

	// 本周新增用户数
	var weekUsers int64
	weekSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE create_time >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)", tableName)
	err = o.Raw(weekSQL).QueryRow(&weekUsers)
	if err == nil {
		stats["week_new_users"] = weekUsers
	}

	// 活跃用户数（最近7天有更新的用户）
	var activeUsers int64
	activeSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE update_time >= DATE_SUB(NOW(), INTERVAL 7 DAY)", tableName)
	err = o.Raw(activeSQL).QueryRow(&activeUsers)
	if err == nil {
		stats["active_users"] = activeUsers
	}

	return stats, nil
}

// getUserDataTableName 获取用户数据表名
func (s *UserService) getUserDataTableName(appId string) string {
	cleanAppId := utils.CleanAppId(appId)
	return fmt.Sprintf("user_data_%s", cleanAppId)
}

// CreateUserDataTable 创建用户数据表（如果不存在）
func (s *UserService) CreateUserDataTable(appId string) error {
	o := orm.NewOrm()

	tableName := s.getUserDataTableName(appId)

	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id VARCHAR(100) NOT NULL,
			data LONGTEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_id (user_id),
			KEY idx_update_time (update_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户数据表'
	`, tableName)

	_, err := o.Raw(createSQL).Exec()
	if err != nil {
		return fmt.Errorf("创建用户数据表失败: %v", err)
	}

	return nil
}
