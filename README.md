# Starring QR Service

扫码核销服务，用于处理扫码验证请求并与后端业务系统交互。本服务为二维码设备提供 HTTP 接口（/api/CheckCode），用于对接业务系统的核销接口。

## 主要功能

- 支持多设备管理和配置
- 签名验证确保安全性
- 自动日志管理和监控
- 支持服务自启动
- 北京时间### 响应
响应格式：`application/json`

#### 成功响应
- 状态码：HTTP 200
- 内容：
  ```json
  {
    "success": true,
    "message": "核销成功"
  }
  ```

#### 失败响应
- 验证失败：
  - 状态码：HTTP 403
  - 内容：
    ```json
    {
      "success": false,
      "message": "核销失败: <具体原因>"
    }
    ```

- 设备未授权：
  - 状态码：HTTP 400
  - 内容：
    ```json
    {
      "success": false,
      "message": "未知设备ID"
    }
    ```

- 系统错误：
  - 状态码：HTTP 500
  - 内容：
    ```json
    {
      "success": false,
      "message": "系统错误"
    }
    ```

### 常见问题排查

1. 设备无法连接服务：
   - 检查网络连接
   - 确认服务运行状态（使用 `service.sh status`）
   - 验证配置的监听地址和端口

2. 核销失败：
   - 检查 monitor.log 中的错误信息
   - 验证业务系统地址是否可访问
   - 确认设备ID是否在配置文件中

3. 日志相关：
   - 所有操作日志在 logs 目录
   - 按日期分文件记录
   - monitor.log 记录异常情况

4. 其他说明：
   - 本服务不包含"开门"逻辑，设备需自行处理
   - 建议定期检查 monitor.log
   - 确保系统时间准确（影响日志和验证）复
- SSH远程管理支持

## 部署说明

### 在 Termux 中部署

1. 解压部署包：
```bash
# 创建应用目录
mkdir -p ~/apps/starring-qr-service
cd ~/apps/starring-qr-service

# 解压文件
tar xzf starring-qr-service-termux.tar.gz

# 设置执行权限
chmod +x starring-qr-service-termux service.sh autostart.sh
```

2. 配置自启动：
```bash
# 创建自启动目录
mkdir -p ~/.termux/boot

# 创建启动脚本链接
ln -sf ~/apps/starring-qr-service/autostart.sh ~/.termux/boot/start-starring-service
```

注意：在 Termux 环境中，服务将安装在以下位置：
- 服务目录：`/data/data/com.termux/files/home/apps/starring-qr-service`
- 日志目录：`/data/data/com.termux/files/home/apps/starring-qr-service/logs`

### 自启动脚本说明

`autostart.sh` 脚本提供以下功能：
- 在设备启动时自动启动服务
- 设置系统时区为北京时间（UTC+8）
- 启动 SSH 服务（如果已安装）
- 自动启动服务监控
- 确保日志目录存在

自启动配置完成后，设备重启时会：
1. 自动设置时区为北京时间
2. 确保必要目录存在（~/apps/starring-qr-service/logs）
3. 启动 starring-qr-service 服务
4. 启动服务监控进程
5. 如果安装了 SSH，会自动启动 SSH 服务

目录结构说明：
```
~/apps/starring-qr-service/
├── starring-qr-service-termux  # 主程序
├── service.sh                  # 服务管理脚本
├── autostart.sh               # 自启动脚本
├── config.yaml               # 配置文件
└── logs/                     # 日志目录
    ├── starring-qr-service-YYYY-MM-DD.log  # 每日服务日志
    ├── monitor.log          # 监控日志
    └── monitor.log.*        # 监控日志轮转文件
```

### 服务管理

使用 `service.sh` 脚本管理服务：

```bash
./service.sh start    # 启动服务（包含监控）
./service.sh stop     # 停止服务
./service.sh restart  # 重启服务
./service.sh status   # 查看状态
./service.sh clean    # 手动清理日志
```

### 日志管理

- 日志位置：`logs/` 目录
- 日志命名：`starring-qr-service-YYYY-MM-DD.log`
- 自动清理：自动保留最近 7 天的日志
- 监控日志：`logs/monitor.log`（记录服务异常和重启信息）
- 时区设置：所有日志使用北京时间（UTC+8）

### 监控功能

服务包含自动监控功能：
- 每 60 秒检查服务状态
- 发现服务异常自动重启
- 记录异常情况到 monitor.log
- 监控日志位于 logs 目录，避免权限问题
- 自动清理和轮转监控日志（超过 1MB 自动轮转）
- 监控状态持久化，避免重复启动

监控进程管理命令：
```bash
./service.sh monitor-start  # 启动监控
./service.sh monitor-stop   # 停止监控
```

监控进程特性：
- 在服务启动时自动开启监控
- 支持单独控制监控进程
- 使用 PID 文件确保只有一个监控实例
- 服务停止时自动清理监控进程
- 异常情况下会尝试强制结束监控进程

### 配置文件

配置文件 `config.yaml` 包含以下配置项：
```yaml
listen_ip: "0.0.0.0"     # 监听地址
port: "8080"             # 监听端口
verify_url: "..."        # 业务系统验证接口
mc_shop_id: 123456       # 商户ID
secret_key: "..."        # 签名密钥
devices:                 # 设备列表
  - id: "device001"      # 设备ID
    name: "前门二维码机"  # 设备名称
```

## 注意事项

1. 确保配置文件中的业务系统地址正确可访问
2. 服务启动时会自动清理7天前的日志
3. 监控日志超过1MB会自动轮转
4. 所有时间均使用北京时间
5. 建议定期检查 monitor.log 了解服务运行状况

## 技术说明

### 工作流程
1. 接收设备发送的二维码内容
2. 验证设备ID的合法性
3. 组装并转发请求到业务系统
4. 处理业务系统响应
5. 返回核销结果给设备

### 安全特性
- 设备白名单管理
- 请求签名验证
- 数据传输加密
- 异常行为记录

## 配置说明
请在项目根目录下创建 `config.yaml`，示例：

```yaml
listen_ip: "0.0.0.0"
port: "8080"
verify_url: "http://your-business-system/verify"
mc_shop_id: 123456
secret_key: "your-secret-key"
devices:
  - id: "device001"
    name: "前门二维码机"
  - id: "device002"
    name: "后门二维码机"
```

## API 使用说明
### 请求
- URL: `/api/CheckCode`
- Method: `POST`
- Content-Type: `application/json`
- Body 示例：
  ```json
  {
    "DeviceID": "device001",
    "CodeVal": "二维码内容"
  }
  ```
  - `DeviceID` 必须为 config.yaml 中已配置的设备 id。

### 响应
- 成功：HTTP 200，body: `核销成功`
- 失败：HTTP 403，body: `核销失败: <原因>`
- 未知设备：HTTP 400，body: `未知设备ID`

## 日志
- 服务启动、请求解析、核销接口调用、核销结果等均有日志输出，包含设备ID和名称。

## 其他
- 本服务不包含“开门”逻辑，设备收到成功响应后自行处理。
