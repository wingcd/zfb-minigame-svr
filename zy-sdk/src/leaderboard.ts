import { Env } from "./env";
import { Http } from "./http";

/**
 * 排行榜
 */
export class Leaderboard {
    /**
     * 提交分数
     * @param type 排行榜类型
     * @param score 分数
     * @returns
     */
    public commitScore(type: string, score: number): Promise<ResponseCommon> {
        return Http.inst.post('/leaderboard/commit', {
            ...Env.getCommonParams(),
            type,
            score,
        }) as any;
    }

    /**
     * 查询排名
     * @param type 排行榜类型
     * @param count 查询数量
     * @param startRank 开始排名名次
     * @returns
     */
    public queryTopRank(type: string, count: number, startRank?: number, test?: boolean): Promise<ResponseCommon & {
        data: {
            type: string,
            count: number,
            list: {
                    playerId: string,
                    score: number,
                    userInfo: {
                        nickName: string,
                        avatarUrl: string,
                    }
            }[]
        }
    }> {
        return Http.inst.post('/leaderboard/queryTopRank', {
            ...Env.getCommonParams(),
            type,
            count,
            test,
            startRank,
        }) as any;
    }
}