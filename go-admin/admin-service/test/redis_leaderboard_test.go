package test

import (
	"admin-service/models"
	"admin-service/services"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// TestRedisLeaderboard Redis排行榜功能测试（直接使用Redis，无MySQL回退）
func TestRedisLeaderboard(t *testing.T) {
	// 初始化数据库和Redis连接
	if err := models.InitDB(); err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}

	if err := models.InitRedis(); err != nil {
		t.Fatalf("初始化Redis失败: %v", err)
	}

	// 创建Redis排行榜服务实例
	redisService := services.NewLeaderboardRedisService()

	// 测试参数
	appId := "test_app"
	leaderboardType := "test_score"

	// 清理测试数据
	defer func() {
		redisService.ClearLeaderboard(appId, leaderboardType)
	}()

	t.Run("测试更新玩家分数", func(t *testing.T) {
		// 测试数据
		testCases := []struct {
			playerId  string
			score     int64
			extraData map[string]interface{}
		}{
			{"player1", 1000, map[string]interface{}{"nickname": "玩家1", "level": 10}},
			{"player2", 800, map[string]interface{}{"nickname": "玩家2", "level": 8}},
			{"player3", 1200, map[string]interface{}{"nickname": "玩家3", "level": 12}},
			{"player4", 900, map[string]interface{}{"nickname": "玩家4", "level": 9}},
			{"player5", 1100, map[string]interface{}{"nickname": "玩家5", "level": 11}},
		}

		// 更新分数
		for _, tc := range testCases {
			err := redisService.UpdateScore(appId, leaderboardType, tc.playerId, tc.score, tc.extraData)
			if err != nil {
				t.Errorf("更新玩家%s分数失败: %v", tc.playerId, err)
			} else {
				t.Logf("成功更新玩家%s分数: %d", tc.playerId, tc.score)
			}
		}
	})

	t.Run("测试获取排行榜数据", func(t *testing.T) {
		// 获取前3名
		data, err := redisService.GetLeaderboard(appId, leaderboardType, 0, 2)
		if err != nil {
			t.Errorf("获取排行榜数据失败: %v", err)
			return
		}

		if len(data) != 3 {
			t.Errorf("期望获取3条数据，实际获取%d条", len(data))
			return
		}

		// 验证排序是否正确（应该是降序）
		for i := 0; i < len(data)-1; i++ {
			if data[i].Score < data[i+1].Score {
				t.Errorf("排序错误: 第%d名分数(%d) < 第%d名分数(%d)", i+1, data[i].Score, i+2, data[i+1].Score)
			}
		}

		// 打印排行榜
		t.Log("排行榜前3名:")
		for i, entry := range data {
			extraDataJson, _ := json.Marshal(entry.ExtraData)
			t.Logf("第%d名: %s (分数: %d, 额外数据: %s)", i+1, entry.PlayerId, entry.Score, string(extraDataJson))
		}
	})

	t.Run("测试获取玩家排名", func(t *testing.T) {
		// 测试各个玩家的排名
		testPlayers := []string{"player1", "player2", "player3", "player4", "player5"}

		for _, playerId := range testPlayers {
			score, rank, err := redisService.GetPlayerRank(appId, leaderboardType, playerId)
			if err != nil {
				t.Errorf("获取玩家%s排名失败: %v", playerId, err)
				continue
			}

			t.Logf("玩家%s: 分数=%d, 排名=%d", playerId, score, rank)

			// 验证排名是否合理
			if rank < 1 || rank > 5 {
				t.Errorf("玩家%s排名异常: %d", playerId, rank)
			}
		}
	})

	t.Run("测试排行榜大小", func(t *testing.T) {
		size, err := redisService.GetLeaderboardSize(appId, leaderboardType)
		if err != nil {
			t.Errorf("获取排行榜大小失败: %v", err)
			return
		}

		expectedSize := int64(5)
		if size != expectedSize {
			t.Errorf("排行榜大小不正确: 期望%d, 实际%d", expectedSize, size)
		} else {
			t.Logf("排行榜大小正确: %d", size)
		}
	})

	t.Run("测试删除玩家", func(t *testing.T) {
		// 删除一个玩家
		playerToDelete := "player2"
		err := redisService.RemovePlayer(appId, leaderboardType, playerToDelete)
		if err != nil {
			t.Errorf("删除玩家%s失败: %v", playerToDelete, err)
			return
		}

		// 验证玩家已被删除
		_, rank, err := redisService.GetPlayerRank(appId, leaderboardType, playerToDelete)
		if err != nil {
			t.Logf("玩家%s已成功删除", playerToDelete)
		} else if rank == 0 {
			t.Logf("玩家%s已不在排行榜中", playerToDelete)
		} else {
			t.Errorf("玩家%s删除失败，仍在排行榜中，排名: %d", playerToDelete, rank)
		}

		// 验证排行榜大小减少
		size, err := redisService.GetLeaderboardSize(appId, leaderboardType)
		if err != nil {
			t.Errorf("获取排行榜大小失败: %v", err)
		} else if size != 4 {
			t.Errorf("删除后排行榜大小不正确: 期望4, 实际%d", size)
		} else {
			t.Log("删除后排行榜大小正确: 4")
		}
	})

	t.Run("测试分数更新策略", func(t *testing.T) {
		// 测试相同玩家的分数更新
		playerId := "player_update_test"

		// 第一次更新
		err := redisService.UpdateScore(appId, leaderboardType, playerId, 500, map[string]interface{}{"test": "first"})
		if err != nil {
			t.Errorf("第一次更新分数失败: %v", err)
			return
		}

		score1, _, _ := redisService.GetPlayerRank(appId, leaderboardType, playerId)
		t.Logf("第一次更新后分数: %d", score1)

		// 等待一秒确保时间戳不同
		time.Sleep(time.Second)

		// 第二次更新（更高分数）
		err = redisService.UpdateScore(appId, leaderboardType, playerId, 600, map[string]interface{}{"test": "second"})
		if err != nil {
			t.Errorf("第二次更新分数失败: %v", err)
			return
		}

		score2, _, _ := redisService.GetPlayerRank(appId, leaderboardType, playerId)
		t.Logf("第二次更新后分数: %d", score2)

		// 验证分数是否正确更新（取决于配置的更新策略）
		if score2 < score1 {
			t.Error("分数更新异常，新分数不应该小于旧分数")
		}

		// 清理测试数据
		redisService.RemovePlayer(appId, leaderboardType, playerId)
	})
}

// BenchmarkRedisLeaderboard Redis排行榜性能测试
func BenchmarkRedisLeaderboard(b *testing.B) {
	// 初始化
	models.InitDB()
	models.InitRedis()

	redisService := services.NewLeaderboardRedisService()
	appId := "benchmark_app"
	leaderboardType := "benchmark_score"

	// 清理
	defer redisService.ClearLeaderboard(appId, leaderboardType)

	b.Run("更新分数性能", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			playerId := fmt.Sprintf("player_%d", i)
			score := int64(i * 10)
			extraData := map[string]interface{}{"iteration": i}

			err := redisService.UpdateScore(appId, leaderboardType, playerId, score, extraData)
			if err != nil {
				b.Errorf("更新分数失败: %v", err)
			}
		}
	})

	// 先插入一些数据用于查询测试
	for i := 0; i < 1000; i++ {
		playerId := fmt.Sprintf("player_%d", i)
		score := int64(i * 10)
		redisService.UpdateScore(appId, leaderboardType, playerId, score, nil)
	}

	b.Run("获取排行榜性能", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := redisService.GetLeaderboard(appId, leaderboardType, 0, 19) // 获取前20名
			if err != nil {
				b.Errorf("获取排行榜失败: %v", err)
			}
		}
	})

	b.Run("获取玩家排名性能", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			playerId := fmt.Sprintf("player_%d", i%1000)
			_, _, err := redisService.GetPlayerRank(appId, leaderboardType, playerId)
			if err != nil {
				b.Errorf("获取玩家排名失败: %v", err)
			}
		}
	})
}

// TestLeaderboardIntegration 集成测试：模拟真实使用场景
func TestLeaderboardIntegration(t *testing.T) {
	// 初始化
	models.InitDB()
	models.InitRedis()

	redisService := services.NewLeaderboardRedisService()
	appId := "integration_app"
	leaderboardType := "weekly_score"

	// 清理
	defer redisService.ClearLeaderboard(appId, leaderboardType)

	t.Run("模拟游戏场景", func(t *testing.T) {
		// 模拟100个玩家的分数更新
		players := make([]string, 100)
		for i := 0; i < 100; i++ {
			players[i] = fmt.Sprintf("player_%03d", i)
		}

		// 随机分数更新
		for round := 1; round <= 5; round++ {
			t.Logf("第%d轮分数更新", round)

			for i, playerId := range players {
				// 模拟随机分数增长
				score := int64((round-1)*100 + i*10 + (i%10)*round)
				extraData := map[string]interface{}{
					"nickname": fmt.Sprintf("玩家%d", i),
					"level":    round*2 + i%5,
					"round":    round,
				}

				err := redisService.UpdateScore(appId, leaderboardType, playerId, score, extraData)
				if err != nil {
					t.Errorf("第%d轮更新玩家%s分数失败: %v", round, playerId, err)
				}
			}

			// 查看当前排行榜前10名
			top10, err := redisService.GetLeaderboard(appId, leaderboardType, 0, 9)
			if err != nil {
				t.Errorf("第%d轮获取排行榜失败: %v", round, err)
				continue
			}

			t.Logf("第%d轮排行榜前3名:", round)
			for i, entry := range top10[:3] {
				t.Logf("  第%d名: %s (分数: %d)", i+1, entry.PlayerId, entry.Score)
			}
		}

		// 最终验证
		finalSize, err := redisService.GetLeaderboardSize(appId, leaderboardType)
		if err != nil {
			t.Errorf("获取最终排行榜大小失败: %v", err)
		} else if finalSize != 100 {
			t.Errorf("最终排行榜大小不正确: 期望100, 实际%d", finalSize)
		} else {
			t.Log("最终排行榜大小正确: 100")
		}

		// 测试排名查询
		testPlayerIds := []string{"player_099", "player_050", "player_001"}
		for _, playerId := range testPlayerIds {
			score, rank, err := redisService.GetPlayerRank(appId, leaderboardType, playerId)
			if err != nil {
				t.Errorf("获取玩家%s排名失败: %v", playerId, err)
			} else {
				t.Logf("玩家%s最终排名: %d (分数: %d)", playerId, rank, score)
			}
		}
	})
}
