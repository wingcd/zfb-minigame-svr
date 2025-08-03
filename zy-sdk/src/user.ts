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
    public login(openId: string): Promise<{
        res: ResponseCommon,
        data: {
            token: string,
            playerId: string,
            isNew: boolean,
            data: any,
        }
    }> {
        let url = '/user/login';
        if(Env.platform === EPlatform.WeChat) {
            url = '/user/login/wx';
        }

        return Http.inst.post(url, {
            appId: Env.appId,
            openId,
        }) as any;
    }    

    public getData(): Promise<{
        res: ResponseCommon,
        data: any
    }> {
        return Http.inst.get('/user/getData') as any;
    }

    public setData(data: any): Promise<{
        res: ResponseCommon,
    }> {
        return Http.inst.post('/user/setData', data) as any;
    }
}