package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

func init() {
	// 确保正确的MIME类型
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".html", "text/html")
}

// 自定义静态文件处理器，确保正确的MIME类型
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	// 移除前缀获取文件路径
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "static/index.html"
	} else if !strings.HasPrefix(path, "static/") {
		path = "static/" + path
	}

	// 设置正确的Content-Type
	ext := filepath.Ext(path)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	http.ServeFile(w, r, path)
}

func main() {
	// 处理静态文件
	http.HandleFunc("/style.css", staticFileHandler)
	http.HandleFunc("/crypto.js", staticFileHandler)
	http.HandleFunc("/api-tester.js", staticFileHandler)

	// 根路径重定向到测试页面
	http.HandleFunc("/", staticFileHandler)

	// API代理接口 - 用于避免跨域问题
	http.HandleFunc("/api/proxy", func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, App-Id, Timestamp, Sign")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 这里可以添加代理逻辑，但现在game-service已经支持CORS，所以可能不需要
		response := map[string]interface{}{
			"message": "Proxy endpoint ready, but game-service should handle CORS directly",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	port := ":8082"
	fmt.Printf("游戏服务测试工具启动在 http://localhost%s\n", port)
	fmt.Println("请确保game-service运行在 http://localhost:8081")

	log.Fatal(http.ListenAndServe(port, nil))
}
