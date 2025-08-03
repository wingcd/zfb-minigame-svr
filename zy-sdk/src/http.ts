import { Env } from "./env";

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
            xhr.onreadystatechange = () => {
                let resp = xhr.responseText || xhr.response;
                if(typeof resp === 'string') {
                    try {
                        resp = JSON.parse(resp);
                    } catch (e) {
                        resp = {
                            code: 500,
                            message: 'response parse error',
                            data: {}
                        };
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
            xhr.send(data);
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

    public async post(url: string, data: any) {
        return await this._requestWithRetry(`${Env.baseUrl}${url}`, 'POST', data);
    }
}