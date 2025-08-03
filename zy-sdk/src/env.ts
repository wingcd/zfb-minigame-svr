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

export class Env {
    public static baseUrl: string = 'http://localhost:3000';
    public static timeout: number = 5000;
    public static appId: string;
    public static playerId: string;
    public static playerName: string;
    public static playerAvatar: string;
    public static token: string;
    public static platform: EPlatform = EPlatform.WeChat;
}