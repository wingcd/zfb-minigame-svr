import { Env } from "./env";
import { GenHash } from "./hash";

export class Http {
    private static _inst: Http;
    public static get inst() {
        if (!this._inst) {
            this._inst = new Http();
        }
        return this._inst;
    }

    private constructor() {
    }

    /**
     * web request
     * @param url 
     * @param method 
     * @param data 
     */
    private _request(url: string, method: string, data: any) {
        return new Promise((resolve, reject) => {
            let xhr = new XMLHttpRequest();
            xhr.open(method, url, true);
            xhr.timeout = Env.timeout;
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.onreadystatechange = () => {
                let resp = xhr.responseText || xhr.response;
                if(typeof resp === 'string') {
                    try {
                        resp = JSON.parse(resp || '{}');
                        if(resp.code == 200) {
                            resp.code = 0;
                        }else if(resp.code) {
                            console.error('response error', resp);
                            reject(resp);
                        }
                    } catch (e) {
                        resp = {
                            code: 500,
                            message: 'response parse error',
                            timespan: Date.now(),
                            data: {}
                        };
                        console.error('response parse error', e, resp);
                        reject(resp);
                    }
                }
                
                if (xhr.readyState === 4) {
                    if (xhr.status === 200) {
                        resolve(resp);
                    } else {
                        reject(resp);
                    }
                }
            };
            let content = data ? JSON.stringify(data) : null;
            xhr.send(content);
        });
    }

    private _requestWithRetry(url: string, method: string, data: any, retry: number = 3) {
        return new Promise(async (resolve, reject) => {
            let i = 0;
            while (i < retry) {
                try {
                    let resp = await this._request(url, method, data);
                    resolve(resp);
                    break;
                } catch (e) {
                    i++;
                }
            }
            reject('retry failed');
        });
    }

    public async get(url: string, ...args: any) {
        let query = '';
        if (args.length > 0) {
            query = '?' + args.map((v: any) => {
                return `${v.key}=${v.value}`;
            }).join('&');
        }
        return await this._requestWithRetry(`${Env.baseUrl}${url}${query}`, 'GET', null);
    }

    public async post(url: string, data: any, checkLogin: boolean = true) {
        if(checkLogin && !Env.isLogined()) {
            return {
                res: {
                    code: 401,
                    message: 'not logined',
                },
                data: {},
            };
        }

        data.sign = GenHash(data);
        return await this._requestWithRetry(`${Env.baseUrl}${url}`, 'POST', data);
    }

    public async ttClickPost(url: string, data: any) {
        return await this._requestWithRetry(`${url}`, 'POST', data);
    }
}