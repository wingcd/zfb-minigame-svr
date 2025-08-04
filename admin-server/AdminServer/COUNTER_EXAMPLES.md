# 计数器功能使用示例

**注意：计数器是绑定到游戏(appId)的，所有玩家共享同一个计数器，不是每个玩家独立的计数器。**

## 快速开始

### 1. 初始化 SDK

```typescript
import { ZYSDK } from '@/zy-sdk';

// 初始化SDK
ZYSDK.init({
    appId: 'your-app-id',
    baseUrl: 'https://your-cloud-function-url'
});
```

### 2. 基础使用

```typescript
// 增加全服计数器
const result = await ZYSDK.counter.incrementCounter('server_events', 1, 'daily');
console.log(`今日全服活动次数: ${result.data.currentValue}`);

// 获取计数器
const counter = await ZYSDK.counter.getCounter('server_events');
console.log(`当前值: ${counter.data.value}`);
```

## 完整游戏示例

### 全服活动系统

```typescript
class ServerEventSystem {
    // 参与全服活动
    async joinServerEvent() {
        try {
            const result = await ZYSDK.counter.incrementDailyChallenge('server_daily_event');
            
            if (result.res.code === 0) {
                const count = result.data.currentValue;
                console.log(`今日全服活动参与次数: ${count}`);
                
                // 检查全服目标
                if (count >= 1000) {
                    this.showServerReward('全服目标达成！');
                } else if (count >= 500) {
                    this.showProgress('全服活动进度过半');
                }
            }
        } catch (error) {
            console.error('活动计数失败:', error);
        }
    }
    
    // 全服BOSS挑战
    async challengeBoss(damage: number) {
        try {
            // 增加挑战次数
            await ZYSDK.counter.incrementDailyChallenge('boss_challenge_count');
            
            // 累计伤害
            await ZYSDK.counter.incrementScore('boss_total_damage', damage);
            
            console.log(`对BOSS造成${damage}伤害`);
        } catch (error) {
            console.error('BOSS挑战计数失败:', error);
        }
    }
    
    // 获取全服进度
    async getServerProgress() {
        try {
            const counters = await ZYSDK.counter.getAllCounters();
            
            if (counters.res.code === 0) {
                console.log('全服数据:');
                counters.data.forEach(counter => {
                    console.log(`${counter.key}: ${counter.value}`);
                });
            }
        } catch (error) {
            console.error('获取进度失败:', error);
        }
    }
    
    private showServerReward(message: string) {
        console.log(`🎉 ${message}`);
        // 显示全服奖励UI
    }
    
    private showProgress(message: string) {
        console.log(`📊 ${message}`);
        // 显示进度UI
    }
}
```

### 全服限时活动系统

```typescript
class ServerEventSystem {
    // 全服限时活动参与（24小时后重置）
    async joinServerEvent(eventId: string) {
        try {
            const result = await ZYSDK.counter.incrementCounter(
                `server_event_${eventId}`, 
                1, 
                'custom', 
                24
            );
            
            if (result.res.code === 0) {
                const count = result.data.currentValue;
                console.log(`全服活动参与次数: ${count}`);
                
                // 检查全服目标
                if (count >= 10000) {
                    console.log('🎉 全服目标达成！所有玩家获得奖励');
                    return true;
                } else if (count >= 5000) {
                    console.log('📊 全服进度过半，继续加油！');
                }
                
                return true;
            }
        } catch (error) {
            console.error('参与活动失败:', error);
            return false;
        }
    }
    
    // 获取全服活动剩余时间
    async getServerEventTimeLeft(eventId: string) {
        try {
            const counter = await ZYSDK.counter.getCounter(`server_event_${eventId}`);
            
            if (counter.res.code === 0 && counter.data.timeToReset) {
                const hours = Math.floor(counter.data.timeToReset / (1000 * 60 * 60));
                const minutes = Math.floor((counter.data.timeToReset % (1000 * 60 * 60)) / (1000 * 60));
                
                console.log(`全服活动剩余时间: ${hours}小时${minutes}分钟`);
                return { 
                    hours, 
                    minutes, 
                    currentProgress: counter.data.value 
                };
            }
        } catch (error) {
            console.error('获取活动时间失败:', error);
        }
        
        return null;
    }
}
```

### 全服周赛系统

```typescript
class ServerWeeklyCompetition {
    // 全服周赛参与
    async joinServerWeeklyBattle() {
        try {
            const result = await ZYSDK.counter.incrementWeeklyBattle('server_weekly_pvp');
            
            if (result.res.code === 0) {
                const count = result.data.currentValue;
                console.log(`本周全服PVP次数: ${count}`);
                
                // 检查全服周赛里程碑
                this.checkServerWeeklyMilestones(count);
                
                return true;
            }
        } catch (error) {
            console.error('周赛参与失败:', error);
            return false;
        }
    }
    
    // 获取全服周赛数据
    async getServerWeeklyData() {
        try {
            const battleCount = await ZYSDK.counter.getWeeklyBattle('server_weekly_pvp');
            const totalDamage = await ZYSDK.counter.getTotalScore('server_weekly_damage');
            
            if (battleCount.res.code === 0 && totalDamage.res.code === 0) {
                return {
                    battles: battleCount.data.value,
                    totalDamage: totalDamage.data.value,
                    resetTime: battleCount.data.resetTime
                };
            }
        } catch (error) {
            console.error('获取周赛数据失败:', error);
        }
        
        return null;
    }
    
    private checkServerWeeklyMilestones(battleCount: number) {
        const milestones = [
            { count: 1000, reward: '全服青铜奖励' },
            { count: 5000, reward: '全服白银奖励' },
            { count: 10000, reward: '全服黄金奖励' },
            { count: 20000, reward: '全服钻石奖励' }
        ];
        
        milestones.forEach(({ count, reward }) => {
            if (battleCount >= count) {
                console.log(`🏆 全服达成${count}场战斗，所有玩家获得${reward}`);
            }
        });
    }
}
```

### 成就系统

```typescript
class AchievementSystem {
    // 检查所有成就
    async checkAchievements() {
        try {
            const allCounters = await ZYSDK.counter.getAllCounters();
            
            if (allCounters.res.code === 0) {
                const achievements = this.calculateAchievements(allCounters.data);
                this.displayAchievements(achievements);
            }
        } catch (error) {
            console.error('检查成就失败:', error);
        }
    }
    
    private calculateAchievements(counters: any[]) {
        const achievements = [];
        
        // 登录成就
        const loginCounter = counters.find(c => c.key === 'daily_login');
        if (loginCounter && loginCounter.value >= 30) {
            achievements.push({
                id: 'login_master',
                name: '登录大师',
                description: '累计登录30天',
                unlocked: true
            });
        }
        
        // 战斗成就
        const battleCounter = counters.find(c => c.key === 'total_battles');
        if (battleCounter && battleCounter.value >= 100) {
            achievements.push({
                id: 'battle_veteran',
                name: '战斗老兵',
                description: '完成100场战斗',
                unlocked: true
            });
        }
        
        // 积分成就
        const scoreCounter = counters.find(c => c.key === 'total_score');
        if (scoreCounter && scoreCounter.value >= 10000) {
            achievements.push({
                id: 'score_legend',
                name: '积分传说',
                description: '累计获得10000分',
                unlocked: true
            });
        }
        
        return achievements;
    }
    
    private displayAchievements(achievements: any[]) {
        console.log('🏆 已解锁成就:');
        achievements.forEach(achievement => {
            if (achievement.unlocked) {
                console.log(`${achievement.name}: ${achievement.description}`);
            }
        });
    }
}
```

### 数据统计面板

```typescript
class StatsDashboard {
    // 显示玩家统计数据
    async showPlayerStats() {
        try {
            const allCounters = await ZYSDK.counter.getAllCounters();
            
            if (allCounters.res.code === 0) {
                console.log('=== 玩家数据统计 ===');
                
                // 按类型分组显示
                const dailyCounters = allCounters.data.filter(c => c.resetType === 'daily');
                const weeklyCounters = allCounters.data.filter(c => c.resetType === 'weekly');
                const permanentCounters = allCounters.data.filter(c => c.resetType === 'permanent');
                
                this.displayCounterGroup('今日数据', dailyCounters);
                this.displayCounterGroup('本周数据', weeklyCounters);
                this.displayCounterGroup('总计数据', permanentCounters);
            }
        } catch (error) {
            console.error('获取统计数据失败:', error);
        }
    }
    
    private displayCounterGroup(title: string, counters: any[]) {
        if (counters.length === 0) return;
        
        console.log(`\n${title}:`);
        counters.forEach(counter => {
            console.log(`  ${counter.key}: ${counter.value}`);
            
            if (counter.timeToReset && counter.timeToReset > 0) {
                const hours = Math.floor(counter.timeToReset / (1000 * 60 * 60));
                console.log(`    重置倒计时: ${hours}小时`);
            }
        });
    }
    
    // 导出数据（用于分析）
    async exportData() {
        try {
            const allCounters = await ZYSDK.counter.getAllCounters();
            
            if (allCounters.res.code === 0) {
                const exportData = {
                    playerId: ZYSDK.env.playerId,
                    exportTime: new Date().toISOString(),
                    counters: allCounters.data
                };
                
                console.log('导出数据:', JSON.stringify(exportData, null, 2));
                return exportData;
            }
        } catch (error) {
            console.error('导出数据失败:', error);
        }
        
        return null;
    }
}
```

## 使用建议

### 1. 错误处理

```typescript
async function safeIncrementCounter(key: string, increment: number = 1) {
    try {
        const result = await ZYSDK.counter.incrementCounter(key, increment);
        
        if (result.res.code === 0) {
            return result.data;
        } else {
            console.error('计数器操作失败:', result.res.message);
            return null;
        }
    } catch (error) {
        console.error('网络错误:', error);
        return null;
    }
}
```

### 2. 缓存优化

```typescript
class CounterCache {
    private cache = new Map<string, any>();
    private cacheTime = new Map<string, number>();
    private CACHE_DURATION = 60000; // 1分钟缓存
    
    async getCounter(key: string, useCache: boolean = true) {
        if (useCache && this.isValidCache(key)) {
            return this.cache.get(key);
        }
        
        const result = await ZYSDK.counter.getCounter(key);
        
        if (result.res.code === 0) {
            this.cache.set(key, result);
            this.cacheTime.set(key, Date.now());
        }
        
        return result;
    }
    
    private isValidCache(key: string): boolean {
        const cacheTime = this.cacheTime.get(key);
        return cacheTime ? (Date.now() - cacheTime) < this.CACHE_DURATION : false;
    }
    
    clearCache() {
        this.cache.clear();
        this.cacheTime.clear();
    }
}
```

### 3. 批量操作

```typescript
async function batchUpdateCounters(updates: Array<{key: string, increment: number}>) {
    const promises = updates.map(update => 
        ZYSDK.counter.incrementCounter(update.key, update.increment)
    );
    
    try {
        const results = await Promise.all(promises);
        return results.filter(result => result.res.code === 0);
    } catch (error) {
        console.error('批量更新失败:', error);
        return [];
    }
}
```

这些示例展示了如何在实际游戏中使用计数器功能，包括每日任务、限时活动、周赛系统、成就系统等常见场景。 