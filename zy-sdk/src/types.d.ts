declare type ResponseCommon = {
    code: number;
    message: string;
};

declare type MailType = 'system' | 'notice' | 'reward';
declare type MailStatus = 'unread' | 'read' | 'received' | 'deleted';
declare type MailAction = 'read' | 'receive' | 'delete';