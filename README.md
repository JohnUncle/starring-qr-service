# starring-qr-service# Starring QR Service

## 简介
本服务为二维码设备提供 HTTP 接口（/api/CheckCode），用于对接业务系统的核销接口，实现二维码核销验证。

## 功能特性
- 提供 `/api/CheckCode` POST 接口，供二维码设备调用。
- 解析请求后，组装并转发到业务系统核销接口。
- 根据业务系统返回结果，返回核销成功或失败。
- 支持通过 `config.yaml` 灵活配置（监听IP、端口、核销URL、门店ID、密钥、多个设备等）。
- 日志记录关键事件与错误，包含设备信息。

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

## 启动方式
```bash
go run main.go
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
