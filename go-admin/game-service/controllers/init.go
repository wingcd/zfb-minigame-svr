package controllers

type CommonRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
}

type CommonResponse struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}
