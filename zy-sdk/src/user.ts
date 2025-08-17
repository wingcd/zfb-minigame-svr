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
        return Http.inst.post('/user/saveData', {
            ...Env.getCommonParams(),
            data,
        }) as any;
    }
}