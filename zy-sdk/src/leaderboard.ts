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
    public commitScore(type: string, score: number): Promise<{
        res: ResponseCommon,
    }> {
        return Http.inst.post('/leaderboard/commit', {
            appId: Env.appId,
            playerId: Env.playerId,
            type,
            score,
            playerInfo: {
                name: Env.playerName,
                avatar: Env.playerAvatar,
            },
        }) as any;
    }

    /**
     * 查询排名
     * @param type 排行榜类型
     * @param count 查询数量
     * @param startRank 开始排名名次
     * @returns
     */
    public queryTopRank(type: string, count: number, startRank?: number): Promise<{
        res: ResponseCommon,
        data: {
            type: string,
            count: number,
            list: {
                    playerId: string,
                    score: number,
                    playerInfo: {
                        name: string,
                        avatar: string,
                    }
            }[]
        }
    }> {
        return Http.inst.get('/leaderboard/queryTopRank', {
            appId: Env.appId,
            type,
            count,
            startRank,
        }) as any;
    }
}