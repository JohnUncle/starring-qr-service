#!/bin/bash

# 等待系统完全启动（防止网络等服务未就绪）
sleep 30

# 启动 SSH 服务
if command -v sshd &> /dev/null; then
    pkill -f sshd  # 确保没有旧的 sshd 进程
    sshd
    echo "SSH service started at $(date)" >> "$HOME/sshd.log"
fi

# 获取脚本所在目录
SERVICE_DIR="$HOME/apps/starring-qr-service"

# 确保服务目录存在
if [ ! -d "$SERVICE_DIR" ]; then
    echo "Service directory not found: $SERVICE_DIR"
    exit 1
fi

# 运行服务启动脚本
$SERVICE_DIR/service.sh start

# 记录启动时间
echo "Auto-start triggered at $(date)" >> "$SERVICE_DIR/logs/autostart.log"
