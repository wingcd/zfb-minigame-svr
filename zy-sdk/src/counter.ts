import { Env } from "./env";
import { Http } from "./http";
import "./types";

/**
 * 计数器功能 - 简化版本，只需要key进行操作
 */
export class Counter {
    /**
     * 增加计数器值
     * @param key 计数器key（需要在后台管理系统中预先创建）
     * @param increment 增加的数量，默认1
     * @param location 点位参数，用于地区排序等，默认为"default"
     * @returns 返回当前计数器值
     */
    public incrementCounter(
        key: string, 
        increment: number = 1,
        location?: string
    ): Promise<{
        res: ResponseCommon,
        data: {
            key: string,
            location: string,
            currentValue: number
        }
    }> {
        const params: any = {
            appId: Env.appId,
            key,
            increment
        };

        if (location) {
            params.location = location;
        }

        return Http.inst.post('/counter/increment', params) as any;
    }

    /**
     * 获取计数器值（返回所有点位）
     * @param key 计数器key（需要在后台管理系统中预先创建）
     * @returns 返回计数器所有点位的值
     */
    public getCounter(key: string): Promise<{
        res: ResponseCommon,
        data: {
            key: string,
            locations: {
                [locationKey: string]: {
                    value: number
                }
            },
            resetType: string,
            resetValue?: number,
            resetTime?: string,
            timeToReset?: number,
            description: string
        }
    }> {
        const params = {
            appId: Env.appId,
            key
        };

        return Http.inst.get('/counter/get', params) as any;
    }

    /**
     * 根据地区获取计数器排行榜
     * @param key 计数器key
     * @returns 返回按值降序排列的地区排行榜
     */
    public async getLocationRanking(key: string): Promise<{
        res: ResponseCommon,
        data: Array<{
            key: string,
            location: string,
            value: number,
            rank: number
        }>
    }> {
        // 获取所有location的数据
        const result = await this.getCounter(key);
        
        if (!result.res || result.res.code !== 0) {
            return result as any;
        }

        // 从新的数据结构中提取locations数据
        const locations = result.data.locations || {};
        
        // 转换为排行榜格式并按值降序排列
        const ranking = Object.entries(locations)
            .map(([locationKey, locationData]) => ({
                key: result.data.key,
                location: locationKey,
                value: locationData.value,
                rank: 0 // 临时值，下面会重新设置
            }))
            .sort((a, b) => b.value - a.value)
            .map((item, index) => ({
                ...item,
                rank: index + 1
            }));

        return {
            res: result.res,
            data: ranking
        };
    }

    /**
     * 获取指定点位的计数器值
     * @param key 计数器key
     * @param location 点位标识
     * @returns 返回指定点位的计数器值
     */
    public async getLocationCounter(key: string, location: string): Promise<{
        res: ResponseCommon,
        data: {
            key: string,
            location: string,
            value: number,
            resetType: string,
            resetValue?: number,
            resetTime?: string,
            timeToReset?: number
        }
    }> {
        const result = await this.getCounter(key);
        
        if (!result.res || result.res.code !== 0) {
            return result as any;
        }

        const locationData = result.data.locations[location];
        if (!locationData) {
                     return {
             res: {
                 code: 4004,
                 message: `计数器[${key}]的点位[${location}]不存在`
             },
             data: null as any
         };
        }

        return {
            res: result.res,
            data: {
                key: result.data.key,
                location: location,
                value: locationData.value,
                resetType: result.data.resetType,
                resetValue: result.data.resetValue,
                resetTime: result.data.resetTime,
                timeToReset: result.data.timeToReset
            }
        };
    }
} 