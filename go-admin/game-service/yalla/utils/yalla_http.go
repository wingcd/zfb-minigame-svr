package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// YallaHTTPClient Yalla HTTP客户端
type YallaHTTPClient struct {
	BaseURL    string
	Timeout    time.Duration
	RetryCount int
	Logger     YallaLogger
}

// YallaLogger 日志接口
type YallaLogger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewYallaHTTPClient 创建HTTP客户端
func NewYallaHTTPClient(baseURL string, timeout int, retryCount int, logger YallaLogger) *YallaHTTPClient {
	return &YallaHTTPClient{
		BaseURL:    baseURL,
		Timeout:    time.Duration(timeout) * time.Second,
		RetryCount: retryCount,
		Logger:     logger,
	}
}

// HTTPResponse HTTP响应结构
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
	Duration   time.Duration
}

// Post 发送POST请求
func (c *YallaHTTPClient) Post(endpoint string, data interface{}, headers map[string]string) (*HTTPResponse, error) {
	return c.request("POST", endpoint, data, headers)
}

// Get 发送GET请求
func (c *YallaHTTPClient) Get(endpoint string, headers map[string]string) (*HTTPResponse, error) {
	return c.request("GET", endpoint, nil, headers)
}

// Put 发送PUT请求
func (c *YallaHTTPClient) Put(endpoint string, data interface{}, headers map[string]string) (*HTTPResponse, error) {
	return c.request("PUT", endpoint, data, headers)
}

// Delete 发送DELETE请求
func (c *YallaHTTPClient) Delete(endpoint string, headers map[string]string) (*HTTPResponse, error) {
	return c.request("DELETE", endpoint, nil, headers)
}

// request 发送HTTP请求
func (c *YallaHTTPClient) request(method, endpoint string, data interface{}, headers map[string]string) (*HTTPResponse, error) {
	url := c.BaseURL + endpoint
	
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("marshal request data failed: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}
	
	var lastErr error
	for i := 0; i <= c.RetryCount; i++ {
		if i > 0 {
			c.Logger.Info("Retrying request", "attempt", i, "url", url)
			time.Sleep(time.Duration(i) * time.Second) // 递增延迟
		}
		
		startTime := time.Now()
		resp, err := c.doRequest(method, url, body, headers)
		duration := time.Since(startTime)
		
		if err != nil {
			lastErr = err
			c.Logger.Error("Request failed", "url", url, "error", err, "attempt", i+1)
			continue
		}
		
		c.Logger.Info("Request completed", "url", url, "status", resp.StatusCode, "duration", duration)
		
		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response body failed: %v", err)
			continue
		}
		
		// 构建响应头map
		respHeaders := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				respHeaders[key] = values[0]
			}
		}
		
		return &HTTPResponse{
			StatusCode: resp.StatusCode,
			Body:       respBody,
			Headers:    respHeaders,
			Duration:   duration,
		}, nil
	}
	
	return nil, fmt.Errorf("request failed after %d attempts: %v", c.RetryCount+1, lastErr)
}

// doRequest 执行HTTP请求
func (c *YallaHTTPClient) doRequest(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	client := &http.Client{
		Timeout: c.Timeout,
	}
	
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}
	
	// 设置默认头部
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "YallaSDK/1.0")
	
	// 设置自定义头部
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	return client.Do(req)
}

// PostJSON 发送JSON POST请求并解析响应
func (c *YallaHTTPClient) PostJSON(endpoint string, requestData interface{}, responseData interface{}) error {
	resp, err := c.Post(endpoint, requestData, nil)
	if err != nil {
		return err
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d, body: %s", resp.StatusCode, string(resp.Body))
	}
	
	if responseData != nil {
		err = json.Unmarshal(resp.Body, responseData)
		if err != nil {
			return fmt.Errorf("unmarshal response failed: %v", err)
		}
	}
	
	return nil
}
