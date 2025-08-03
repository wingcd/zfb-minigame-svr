import { Env } from "./env";
import { Http } from "./http";
import { Leaderboard } from "./leaderboard";
import { User } from "./user";

type ZYSDKOptions = {
    appId: string,
    baseUrl?: string,
    timeout?: number,
};

class _ZYSDK {
    private static _inst: _ZYSDK;
    public static get inst() {
        if (!this._inst) {
            throw new Error('ZYSDK not initialized');
        }
        return this._inst;
    }

    public readonly http = Http.inst;
    public readonly env = Env;

    public readonly user = new User();
    public readonly leaderboard = new Leaderboard();

    public init(opts: ZYSDKOptions) {
        Env.baseUrl = opts.baseUrl || 'https://env-00jxt0uhcb2h.dev-hz.cloudbasefunction.cn';
        Env.timeout = opts.timeout || 5000;
        Env.appId = opts.appId;
    }
}

export const ZYSDK = new _ZYSDK();