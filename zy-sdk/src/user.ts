import { Env, EPlatform } from "./env";
import { Http } from "./http";

/**
 * 用户
 */
export class User {
    /**
     * login
     * @param openId 
     * @param unionId 
     * @returns 
     */
    public login(code: string): Promise<ResponseCommon & {
        data: {
            token: string,
            playerId: string,
            isNew: boolean,
            openId: string,
            unionid: string,
            data: any,
        }
    }> {
        let url = '/user/login';
        if(Env.platform === EPlatform.WeChat) {
            url = '/user/login/wx';
        }

        return Http.inst.post(url, {
            ...Env.getCommonParams(),
            code,
        }, false) as any;
    }    

    public getData(): Promise<ResponseCommon & {
        data: any
    }> {
        return Http.inst.post('/user/getData', {
            ...Env.getCommonParams(),
        }) as any;
    }

    public saveData(data: any): Promise<ResponseCommon> {
        const requestData: any = {
            ...Env.getCommonParams(),
            data,
        };
            
        return Http.inst.post('/user/saveData', requestData) as any;
    }

    /**
     * 保存玩家基本信息
     * @param userInfo 用户信息对象
     * @returns Promise<ResponseCommon>
     */
    public saveUserInfo(userInfo: {
        nickName?: string;
        avatarUrl?: string;
        gender?: number; // 0-未知, 1-男, 2-女
        province?: string;
        city?: string;
        level?: number;
        exp?: number;
        lastLoginTime?: string;
        lastLogoutTime?: string;
        lastLoginIp?: string;
        lastLogoutIp?: string;
        lastLoginDevice?: string;
        lastLogoutDevice?: string;  
    }): Promise<ResponseCommon> {
        return Http.inst.post('/user/saveUserInfo', {
            ...Env.getCommonParams(),
            userInfo: JSON.stringify(userInfo),
        }) as any;
    }
}