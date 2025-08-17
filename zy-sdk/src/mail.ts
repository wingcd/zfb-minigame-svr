import { Http } from "./http";
import { Env } from "./env";

export enum MailType {
    System = 'system',
    Notice = 'notice',
    Reward = 'reward',
}

export enum MailStatus {
    Unread = 'unread',
    Read = 'read',
    Received = 'received',
    Deleted = 'deleted',
}

export enum MailAction {
    Read = 'read',
    Receive = 'receive',
    Delete = 'delete',
}

export interface MailReward {
    type: string;
    name: string;
    amount: number;
    description?: string;
}

export interface MailInfo {
    mailId: string;
    title: string;
    content: string;
    type: MailType;
    rewards?: MailReward[];
    publishTime: string;
    expireTime?: string;
    isRead: boolean;
    isReceived: boolean;
    isDeleted: boolean;
    status: MailStatus;
}

export interface GetMailsParams {
    page?: number;
    pageSize?: number;
    type?: MailType;
    status?: MailStatus;
}

export interface GetMailsResponse {
    code: number;
    msg: string;
    timestamp: number;
    data: {
        list: MailInfo[];
        total: number;
        page: number;
        pageSize: number;
        totalPages: number;
        hasMore: boolean;
    };
}

export interface UpdateMailStatusParams {
    appId: string;
    playerId: string;
    mailId: string;
    action: MailAction;
}

export interface UpdateMailStatusResponse {
    code: number;
    msg: string;
    timestamp: number;
    data?: {
        rewards?: MailReward[];
    };
}

export interface GetUnreadCountParams {
    openId: string;
}

export interface GetUnreadCountResponse {
    code: number;
    msg: string;
    timestamp: number;
    data: {
        unreadCount: number;
        unreceiveCount: number;
    };
}

export class Mail {
    /**
     * 获取用户邮件列表
     * @param params 查询参数
     * @returns 邮件列表响应
     */
    public async getMails(params: GetMailsParams): Promise<GetMailsResponse> {
        const queryParams: any = {
            ...Env.getCommonParams(),
            page: params.page || 1,
            pageSize: params.pageSize || 20
        };

        if (params.type) {
            queryParams.type = params.type;
        }
        if (params.status) {
            queryParams.status = params.status;
        }

        return Http.inst.get('/mail/getUserMails', ...queryParams) as Promise<GetMailsResponse>;
    }

    /**
     * 更新邮件状态（阅读、领取奖励、删除）
     * @param params 更新参数
     * @returns 更新结果
     */
    public async updateStatus(params: UpdateMailStatusParams): Promise<UpdateMailStatusResponse> {
        const requestData = {
            ...Env.getCommonParams(),
            mailId: params.mailId,
            action: params.action
        };

        return Http.inst.post('/mail/updateStatus', requestData) as Promise<UpdateMailStatusResponse>;
    }

    /**
     * 阅读邮件
     * @param openId 用户openId
     * @param mailId 邮件ID
     * @returns 操作结果
     */
    public async readMail(mailId: string): Promise<UpdateMailStatusResponse> {
        return this.updateStatus({
            ...Env.getCommonParams(),
            mailId,
            action: MailAction.Read
        });
    }

    /**
     * 领取邮件奖励
     * @param mailId 邮件ID
     * @returns 领取结果，包含奖励信息
     */
    public async receiveMail(mailId: string): Promise<UpdateMailStatusResponse> {
        return this.updateStatus({
            ...Env.getCommonParams(),
            mailId,
            action: MailAction.Receive
        });
    }

    /**
     * 删除邮件
     * @param openId 用户openId
     * @param mailId 邮件ID
     * @returns 操作结果
     */
    public async deleteMail(mailId: string): Promise<UpdateMailStatusResponse> {
        return this.updateStatus({
            ...Env.getCommonParams(),
            mailId,
            action: MailAction.Delete
        });
    }
} 