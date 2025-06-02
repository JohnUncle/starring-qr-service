package main

import (
	"log"
	"net/http"
)

func main() {
	InitLogger("logs")
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	http.HandleFunc("/api/CheckCode", CheckCodeHandler(cfg))
	http.HandleFunc("/api/IsConnect", IsConnectHandler(cfg))
	addr := cfg.ListenIP + ":" + cfg.Port
	log.Printf("--------------------")
	log.Printf("服务正在启动...")
	log.Printf("监听地址: %s", addr)
	log.Printf("配置信息:")
	log.Printf("- Verify URL: %s", cfg.VerifyURL)
	log.Printf("- McShop ID: %d", cfg.McShopID)
	log.Printf("- 设备数量: %d", len(cfg.Devices))
	log.Printf("--------------------")
	log.Printf("服务已启动，等待请求...")
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
