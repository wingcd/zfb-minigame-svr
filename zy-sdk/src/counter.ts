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
     * @returns 返回当前计数器值
     */
    public incrementCounter(
        key: string, 
        increment: number = 1,
    ): Promise<{
        res: ResponseCommon,
        data: {
            key: string,
            currentValue: number
        }
    }> {
        const params: any = {
            appId: Env.appId,
            key,
            increment
        };

        return Http.inst.post('/counter/increment', params) as any;
    }

    /**
     * 获取计数器值
     * @param key 计数器key（需要在后台管理系统中预先创建）
     * @returns 返回计数器当前值
     */
    public getCounter(key: string): Promise<{
        res: ResponseCommon,
        data: {
            key: string,
            value: number
        }
    }> {
        const params: any = {
            appId: Env.appId,
            key
        };

        return Http.inst.get('/counter/get', params) as any;
    }
} 