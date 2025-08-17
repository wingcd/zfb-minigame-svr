import { Env, EPlatform } from "./env";
import { Http } from "./http";
import { Leaderboard } from "./leaderboard";
import { User } from "./user";
import { Counter } from "./counter";
import { Mail } from "./mail";

type ZYSDKOptions = {
    platform: EPlatform,
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
    public readonly counter = new Counter();
    public readonly mail = new Mail();

    public init(opts: ZYSDKOptions) {
        Env.init(opts);
    }
}

export const ZYSDK = new _ZYSDK();