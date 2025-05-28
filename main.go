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
	http.HandleFunc("/api/QueryCmd", QueryCmdHandler(cfg))
	addr := cfg.ListenIP + ":" + cfg.Port
	log.Printf("服务启动，监听: %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
