# 计数器功能使用示例

**注意：计数器是绑定到游戏(appId)的，所有玩家共享同一个计数器，不是每个玩家独立的计数器。**

**新增功能：现在支持点位参数(location)，一个计数器可以记录不同点位的值，可用于地区排序、服务器排行等功能。**

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
// 增加全服计数器（使用默认点位）
const result = await ZYSDK.counter.incrementCounter('server_events', 1);
console.log(`全服活动次数: ${result.data.currentValue}`);

// 增加指定地区计数器
const beijingResult = await ZYSDK.counter.incrementCounter('server_events', 1, 'beijing');
console.log(`北京地区活动次数: ${beijingResult.data.currentValue}`);

// 获取计数器所有点位数据
const allCounters = await ZYSDK.counter.getCounter('server_events');
console.log('所有地区数据:', allCounters.data);
console.log('北京地区值:', allCounters.data.locations.beijing?.value || 0);
console.log('默认点位值:', allCounters.data.locations.default?.value || 0);

// 获取指定点位的计数器值（新增方法）
const beijingCounter = await ZYSDK.counter.getLocationCounter('server_events', 'beijing');
console.log(`北京当前值: ${beijingCounter.data.value}`);
```

## 完整游戏示例

### 1. 地区竞赛系统

```typescript
class RegionCompetitionSystem {
    // 参与地区活动
    async participateActivity(region: string, participantCount: number = 1) {
        try {
            const result = await ZYSDK.counter.incrementCounter(
                'region_competition', 
                participantCount, 
                region
            );
            
            console.log(`${region}地区新增参与者${participantCount}人，总计：${result.data.currentValue}人`);
            return result;
        } catch (error) {
            console.error('参与活动失败:', error);
        }
    }
    
    // 获取地区排行榜
    async getRegionRanking() {
        try {
            const ranking = await ZYSDK.counter.getLocationRanking('region_competition');
            
            console.log('=== 地区竞赛排行榜 ===');
            ranking.data.forEach(region => {
                console.log(`第${region.rank}名: ${region.location} - ${region.value}人参与`);
            });
            
            return ranking.data;
        } catch (error) {
            console.error('获取排行榜失败:', error);
        }
    }
    
    // 获取特定地区排名
    async getRegionRank(targetRegion: string) {
        const ranking = await this.getRegionRanking();
        const regionData = ranking?.find(r => r.location === targetRegion);
        
        if (regionData) {
            console.log(`${targetRegion}地区排名第${regionData.rank}位，共${regionData.value}人参与`);
            return regionData;
        } else {
            console.log(`${targetRegion}地区暂无参与数据`);
            return null;
        }
    }
}

// 使用示例
const competition = new RegionCompetitionSystem();

// 模拟各地区参与活动
await competition.participateActivity('beijing', 15);
await competition.participateActivity('shanghai', 20);
await competition.participateActivity('guangzhou', 10);
await competition.participateActivity('shenzhen', 25);

// 查看排行榜
await competition.getRegionRanking();

// 查看特定地区排名
await competition.getRegionRank('beijing');
```

### 2. 服务器统计系统

```typescript
class ServerStatsSystem {
    // 玩家上线统计
    async playerOnline(serverId: string) {
        const result = await ZYSDK.counter.incrementCounter('online_players', 1, serverId);
        console.log(`服务器${serverId}在线玩家数: ${result.data.currentValue}`);
        return result;
    }
    
    // 玩家下线统计（可以用负数减少）
    async playerOffline(serverId: string) {
        const result = await ZYSDK.counter.incrementCounter('online_players', -1, serverId);
        console.log(`服务器${serverId}在线玩家数: ${result.data.currentValue}`);
        return result;
    }
    
    // 获取所有服务器状态
    async getAllServerStats() {
        const stats = await ZYSDK.counter.getCounter('online_players');
        
        console.log('=== 服务器在线统计 ===');
        Object.entries(stats.data.locations).forEach(([serverId, data]) => {
            console.log(`${serverId}: ${data.value}人在线`);
        });
        
        return stats.data;
    }
    
    // 获取服务器排行
    async getServerRanking() {
        const ranking = await ZYSDK.counter.getLocationRanking('online_players');
        
        console.log('=== 服务器人气排行 ===');
        ranking.data.forEach(server => {
            console.log(`第${server.rank}名: ${server.location} - ${server.value}人在线`);
        });
        
        return ranking.data;
    }
}

// 使用示例
const serverStats = new ServerStatsSystem();

// 模拟玩家上线
await serverStats.playerOnline('server_01');
await serverStats.playerOnline('server_02');
await serverStats.playerOnline('server_01'); // 服务器1再上线一个玩家

// 查看所有服务器状态
await serverStats.getAllServerStats();

// 查看服务器排行
await serverStats.getServerRanking();
```

### 3. 多维度活动统计

```typescript
class MultiDimensionEventSystem {
    // 按活动类型和地区统计
    async recordEvent(eventType: string, region: string, count: number = 1) {
        const result = await ZYSDK.counter.incrementCounter(eventType, count, region);
        console.log(`${region}地区${eventType}活动: +${count}, 总计: ${result.data.currentValue}`);
        return result;
    }
    
    // 获取特定活动的地区排行
    async getEventRegionRanking(eventType: string) {
        const ranking = await ZYSDK.counter.getLocationRanking(eventType);
        
        console.log(`=== ${eventType}活动地区排行 ===`);
        ranking.data.forEach(region => {
            console.log(`${region.rank}. ${region.location}: ${region.value}次`);
        });
        
        return ranking.data;
    }
    
    // 比较不同活动的热度
    async compareEventPopularity(eventTypes: string[]) {
        const eventStats = [];
        
        for (const eventType of eventTypes) {
            const data = await ZYSDK.counter.getCounter(eventType);
            const totalCount = Object.values(data.data.locations)
                .reduce((sum, location) => sum + location.value, 0);
            
            eventStats.push({
                eventType,
                totalParticipants: totalCount
            });
        }
        
        // 按参与人数排序
        eventStats.sort((a, b) => b.totalParticipants - a.totalParticipants);
        
        console.log('=== 活动热度排行 ===');
        eventStats.forEach((event, index) => {
            console.log(`${index + 1}. ${event.eventType}: ${event.totalParticipants}人参与`);
        });
        
        return eventStats;
    }
}

// 使用示例
const eventSystem = new MultiDimensionEventSystem();

// 记录不同类型活动的参与情况
await eventSystem.recordEvent('pvp_battle', 'beijing', 5);
await eventSystem.recordEvent('pvp_battle', 'shanghai', 8);
await eventSystem.recordEvent('pvp_battle', 'guangzhou', 3);

await eventSystem.recordEvent('pve_dungeon', 'beijing', 12);
await eventSystem.recordEvent('pve_dungeon', 'shanghai', 15);
await eventSystem.recordEvent('pve_dungeon', 'guangzhou', 6);

await eventSystem.recordEvent('guild_war', 'beijing', 20);
await eventSystem.recordEvent('guild_war', 'shanghai', 18);

// 查看各活动的地区排行
await eventSystem.getEventRegionRanking('pvp_battle');
await eventSystem.getEventRegionRanking('pve_dungeon');

// 比较活动热度
await eventSystem.compareEventPopularity(['pvp_battle', 'pve_dungeon', 'guild_war']);
```

### 4. 传统计数器（向后兼容）

```typescript
class TraditionalCounterSystem {
    // 全服统一计数器（不使用地区）
    async incrementGlobalEvent(eventName: string, count: number = 1) {
        // 不传location参数，使用默认点位
        const result = await ZYSDK.counter.incrementCounter(eventName, count);
        console.log(`全服${eventName}: ${result.data.currentValue}`);
        return result;
    }
    
    // 获取全服数据
    async getGlobalEvent(eventName: string) {
        // 获取所有数据，然后取默认点位的值
        const result = await ZYSDK.counter.getCounter(eventName);
        const defaultValue = result.data.locations.default?.value || 0;
        console.log(`全服${eventName}当前值: ${defaultValue}`);
        return { ...result, data: { ...result.data, value: defaultValue } };
    }
}

// 使用示例
const globalSystem = new TraditionalCounterSystem();

// 全服活动计数
await globalSystem.incrementGlobalEvent('world_boss_killed', 1);
await globalSystem.incrementGlobalEvent('total_logins', 100);

// 查看全服数据
await globalSystem.getGlobalEvent('world_boss_killed');
await globalSystem.getGlobalEvent('total_logins');
```

## 进阶使用技巧

### 1. 动态创建点位

```typescript
// 系统会自动为新的location创建记录
async function createNewServerLocation(newServerId: string) {
    // 第一次访问新location时，系统会基于default配置自动创建
    await ZYSDK.counter.incrementCounter('player_count', 1, newServerId);
    console.log(`新服务器${newServerId}已创建计数器`);
}

// 为新开服务器创建统计
await createNewServerLocation('server_03');
```

### 2. 批量统计分析

```typescript
async function analyzeLocationPerformance(counterKey: string) {
    const allData = await ZYSDK.counter.getCounter(counterKey);
    
    const locations = Object.entries(allData.data.locations);
    const totalValue = locations.reduce((sum, [key, data]) => sum + data.value, 0);
    const avgValue = totalValue / locations.length;
    
    console.log(`=== ${counterKey} 统计分析 ===`);
    console.log(`总计: ${totalValue}`);
    console.log(`平均值: ${avgValue.toFixed(2)}`);
    console.log(`点位数量: ${locations.length}`);
    
    // 找出表现最好和最差的点位
    const sortedLocations = locations.sort((a, b) => b[1].value - a[1].value);
    const bestLocation = sortedLocations[0];
    const worstLocation = sortedLocations[sortedLocations.length - 1];
    
    console.log(`表现最好: ${bestLocation[0]} (${bestLocation[1].value})`);
    console.log(`表现最差: ${worstLocation[0]} (${worstLocation[1].value})`);
    
    return {
        total: totalValue,
        average: avgValue,
        best: { location: bestLocation[0], value: bestLocation[1].value },
        worst: { location: worstLocation[0], value: worstLocation[1].value },
        locations: allData.data.locations
    };
}

// 分析地区活动表现
await analyzeLocationPerformance('region_events');
```

### 3. 实时监控系统

```typescript
class RealTimeMonitorSystem {
    private monitoringActive = false;
    
    // 开始监控
    async startMonitoring(counterKey: string, interval: number = 30000) {
        this.monitoringActive = true;
        console.log(`开始监控 ${counterKey}，间隔 ${interval}ms`);
        
        while (this.monitoringActive) {
            await this.checkAndReport(counterKey);
            await this.sleep(interval);
        }
    }
    
    // 停止监控
    stopMonitoring() {
        this.monitoringActive = false;
        console.log('监控已停止');
    }
    
    private async checkAndReport(counterKey: string) {
        try {
            const ranking = await ZYSDK.counter.getLocationRanking(counterKey);
            
            console.log(`[${new Date().toLocaleTimeString()}] ${counterKey} 实时排行:`);
            ranking.data.slice(0, 3).forEach(loc => {
                console.log(`  ${loc.rank}. ${loc.location}: ${loc.value}`);
            });
            
            // 可以在这里添加警报逻辑
            const topLocation = ranking.data[0];
            if (topLocation && topLocation.value > 1000) {
                console.log(`⚠️ 警报: ${topLocation.location} 数值过高 (${topLocation.value})`);
            }
        } catch (error) {
            console.error('监控检查失败:', error);
        }
    }
    
    private sleep(ms: number): Promise<void> {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

// 使用示例
const monitor = new RealTimeMonitorSystem();

// 开始监控（在实际应用中，这应该在后台运行）
// await monitor.startMonitoring('server_load', 60000); // 每分钟检查一次

// 停止监控
// monitor.stopMonitoring();
```

## 最佳实践

1. **合理设计点位**: 点位名称应该有意义且易于理解，如使用城市名、服务器ID等
2. **错误处理**: 始终包含适当的错误处理逻辑
3. **性能考虑**: 对于高频操作，考虑批量更新或缓存机制
4. **数据分析**: 定期分析计数器数据，为游戏运营提供决策依据
5. **监控报警**: 为关键指标设置监控和报警机制

```typescript
// 错误处理示例
async function safeIncrementCounter(key: string, increment: number, location?: string) {
    try {
        const result = await ZYSDK.counter.incrementCounter(key, increment, location);
        return { success: true, data: result.data };
    } catch (error) {
        console.error(`计数器操作失败 - Key: ${key}, Location: ${location}`, error);
        return { success: false, error: error.message };
    }
}

// 使用安全包装函数
const result = await safeIncrementCounter('user_actions', 1, 'beijing');
if (result.success) {
    console.log('操作成功:', result.data);
} else {
    console.log('操作失败:', result.error);
}
``` 

这些示例展示了如何在实际游戏中使用计数器功能，包括每日任务、限时活动、周赛系统、成就系统等常见场景。 