import { Http } from "./http";
import { Env } from "./env";

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
    type: 'system' | 'notice' | 'reward';
    rewards?: MailReward[];
    publishTime: string;
    expireTime?: string;
    isRead: boolean;
    isReceived: boolean;
    isDeleted: boolean;
    status: 'unread' | 'read' | 'received' | 'deleted';
}

export interface GetMailsParams {
    openId: string;
    page?: number;
    pageSize?: number;
    type?: 'system' | 'notice' | 'reward';
    status?: 'unread' | 'read' | 'received' | 'deleted';
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
    action: 'read' | 'receive' | 'delete';
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
        const queryParams = [
            { key: 'appId', value: Env.appId },
            { key: 'openId', value: params.openId },
            { key: 'page', value: params.page || 1 },
            { key: 'pageSize', value: params.pageSize || 20 }
        ];

        if (params.type) {
            queryParams.push({ key: 'type', value: params.type });
        }
        if (params.status) {
            queryParams.push({ key: 'status', value: params.status });
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
            appId: Env.appId,
            playerId: params.playerId,
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
            appId: Env.appId,
            playerId: Env.playerId,
            mailId,
            action: 'read'
        });
    }

    /**
     * 领取邮件奖励
     * @param mailId 邮件ID
     * @returns 领取结果，包含奖励信息
     */
    public async receiveMail(mailId: string): Promise<UpdateMailStatusResponse> {
        return this.updateStatus({
            appId: Env.appId,
            playerId: Env.playerId,
            mailId,
            action: 'receive'
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
            appId: Env.appId,
            playerId: Env.playerId,
            mailId,
            action: 'delete'
        });
    }
} 