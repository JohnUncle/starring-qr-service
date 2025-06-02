#!/bin/bash

SERVICE_NAME="starring-qr-service"
# 检测是否在 Termux 环境中
if [ -d "/data/data/com.termux/files/home" ]; then
    BASE_DIR="/data/data/com.termux/files/home/apps"
else
    BASE_DIR="$HOME"
fi

SERVICE_DIR="$BASE_DIR/$SERVICE_NAME"
EXEC_NAME="starring-qr-service"
LOG_DIR="$SERVICE_DIR/logs"
LOG_DATE=$(date +%Y-%m-%d)
LOG_FILE="$LOG_DIR/${SERVICE_NAME}-${LOG_DATE}.log"
MONITOR_LOG="$LOG_DIR/monitor.log"
DAYS_TO_KEEP_LOGS=7  # 保留最近7天的日志
MONITOR_INTERVAL=60   # 监控间隔（秒）
MONITOR_PID_FILE="$LOG_DIR/${SERVICE_NAME}_monitor.pid"  # 将PID文件放在logs目录下

check_status() {
    if pgrep -f "$EXEC_NAME" > /dev/null; then
        return 0
    else
        return 1
    fi
}

start_service() {
    if check_status; then
        echo "[$SERVICE_NAME] is already running"
        return 1
    fi
    
    cd "$SERVICE_DIR"
    mkdir -p "$LOG_DIR"
    
    # 清理旧日志
    clean_old_logs
    
    # 设置时区为北京时间并启动服务
    export TZ=Asia/Shanghai
    chmod +x $EXEC_NAME
    ./$EXEC_NAME > "$LOG_FILE" 2>&1 &
    sleep 1
    
    if check_status; then
        echo "[$SERVICE_NAME] started successfully at $(date)"
        start_monitor
    else
        echo "[$SERVICE_NAME] failed to start"
        return 1
    fi
}

stop_service() {
    if ! check_status; then
        echo "[$SERVICE_NAME] is not running"
        return 1
    fi
    
    # 先停止监控
    stop_monitor
    
    pkill -f "$EXEC_NAME"
    sleep 1
    
    if ! check_status; then
        echo "[$SERVICE_NAME] stopped successfully at $(date)"
    else
        echo "[$SERVICE_NAME] failed to stop"
        return 1
    fi
}

clean_old_logs() {
    echo "Cleaning logs older than $DAYS_TO_KEEP_LOGS days..."
    
    # 清理服务日志
    find "$LOG_DIR" -type f -name "${SERVICE_NAME}-*.log" -mtime +$DAYS_TO_KEEP_LOGS -exec rm {} \;
    
    # 清理旧的监控日志
    find "$LOG_DIR" -type f -name "monitor.log.*" -mtime +$DAYS_TO_KEEP_LOGS -exec rm {} \;
    
    # 如果当前监控日志超过1MB，进行轮转
    if [ -f "$MONITOR_LOG" ]; then
        LOG_SIZE=$(stat -f "%s" "$MONITOR_LOG" 2>/dev/null || wc -c < "$MONITOR_LOG" 2>/dev/null)
        if [ -n "$LOG_SIZE" ] && [ "$LOG_SIZE" -gt 1048576 ]; then
            mv "$MONITOR_LOG" "$MONITOR_LOG.$(date +%Y%m%d_%H%M%S).old"
        fi
    fi
    
    echo "Found and cleaned these old log files:"
    find "$LOG_DIR" -type f -name "${SERVICE_NAME}-*.log" -mtime +$DAYS_TO_KEEP_LOGS -ls
    
    echo "Log cleanup completed"
}

start_monitor() {
    if [ -f "$MONITOR_PID_FILE" ]; then
        local pid=$(cat "$MONITOR_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo "Monitor is already running"
            return
        fi
    fi
    
    # 启动监控进程
    (
        while true; do
            # 检查并轮转监控日志
            if [ -f "$MONITOR_LOG" ]; then
                LOG_SIZE=$(stat -f "%s" "$MONITOR_LOG" 2>/dev/null || wc -c < "$MONITOR_LOG" 2>/dev/null)
                if [ -n "$LOG_SIZE" ] && [ "$LOG_SIZE" -gt 1048576 ]; then
                    mv "$MONITOR_LOG" "$MONITOR_LOG.$(date +%Y%m%d_%H%M%S).old"
                    echo "[$(date)] Monitor log rotated due to size > 1MB" >> "$MONITOR_LOG"
                fi
            fi
            
            # 检查服务状态
            if ! check_status; then
                echo "[$(date)] Service crashed, restarting..." >> "$MONITOR_LOG"
                start_service
            fi
            sleep $MONITOR_INTERVAL
        done
    ) &
    
    echo $! > "$MONITOR_PID_FILE"
    echo "Service monitor started"
}

stop_monitor() {
    if [ -f "$MONITOR_PID_FILE" ]; then
        local pid=$(cat "$MONITOR_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo "Stopping monitor process (PID: $pid)..."
            kill "$pid"
            # 等待进程真正结束
            for i in {1..5}; do
                if ! kill -0 "$pid" 2>/dev/null; then
                    break
                fi
                sleep 1
            done
            # 如果进程还在运行，强制终止
            if kill -0 "$pid" 2>/dev/null; then
                echo "Monitor process not responding, force killing..."
                kill -9 "$pid"
            fi
        else
            echo "Monitor process is not running"
        fi
        rm -f "$MONITOR_PID_FILE"
        echo "Service monitor stopped"
    else
        echo "No monitor PID file found"
    fi
}

status_service() {
    if check_status; then
        echo "[$SERVICE_NAME] is running"
        if [ -f "$MONITOR_PID_FILE" ]; then
            echo "Service monitor is active"
        else
            echo "Service monitor is not active"
        fi
        echo "Recent logs:"
        tail -n 5 "$LOG_FILE"
    else
        echo "[$SERVICE_NAME] is not running"
    fi
}

case "$1" in
    start)
        start_service
        ;;
    stop)
        stop_service
        ;;
    restart)
        stop_service
        start_service
        ;;
    status)
        status_service
        ;;
    clean)
        clean_old_logs
        ;;
    monitor-start)
        start_monitor
        ;;
    monitor-stop)
        stop_monitor
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|clean|monitor-start|monitor-stop}"
        exit 1
        ;;
esac
