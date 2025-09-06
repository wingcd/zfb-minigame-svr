package main

import (
	"admin-service/utils"
	"fmt"
	"log"
	"os"
)

func main() {
	log.Println("开始强制重新安装数据库...")

	// 删除安装锁文件
	if err := os.Remove(".installed"); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("删除安装锁文件失败: %v", err)
		}
	}

	// 执行重新安装
	if err := utils.AutoInstall(); err != nil {
		log.Fatalf("重新安装失败: %v", err)
	}

	fmt.Println("数据库重新安装完成！")
}
