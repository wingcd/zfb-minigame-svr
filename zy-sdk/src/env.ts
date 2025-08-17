import { GenHash } from "./hash";

export enum EPlatform {
    Web,
    WeChat,
    Alipay,
    Taobao,
    QQ,
    Oppo,
    Vivo,
    ByteDance,
    BaiDu,
    Mi,
    Huawei,
}

type ZYSDKOptions = {
    platform: EPlatform,
    appId: string,
    baseUrl?: string,
    timeout?: number,
};

export class Env {
    public static initialized: boolean = false;
    public static version: string = '1.0.0';
    public static baseUrl: string = 'https://env-00jxt0uhcb2h.dev-hz.cloudbasefunction.cn';
    public static timeout: number = 5000;
    public static appId: string;
    public static openId: string;
    public static playerId: string;
    public static playerName: string;
    public static playerAvatar: string;
    public static token: string;
    public static platform: EPlatform = EPlatform.Web;

    public static init(cfg:ZYSDKOptions) {
        this.platform = cfg.platform;
        this.appId = cfg.appId;
        this.baseUrl = cfg.baseUrl || this.baseUrl;
        this.timeout = cfg.timeout || this.timeout;
        this.initialized = true;
    }

    public static isLogined() {
        return !!this.playerId;
    }

    public static getCommonParams() {
        return {
            appId: this.appId,
            playerId: this.playerId || '',
            token: this.token || '',
            timestamp: Date.now(),
            ver: this.version,
        };
    }
}