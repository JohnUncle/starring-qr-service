package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type QRRequest struct {
	CodeVal   string `json:"CodeVal"`
	CodeType  string `json:"CodeType"`
	BrushTime string `json:"BrushTime"`
	ViewId    string `json:"ViewId"`
	UID       string `json:"UID"`
	UKey      string `json:"UKey"`
	SN        string `json:"SN"`
	IsOnline  string `json:"IsOnline"`
	Property  string `json:"Property"`
	Timestamp string `json:"Timestamp"`
	Sign      string `json:"Sign"`
}

type VerifyResponse struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
}

// QueryCmdRequest 请求结构体
// 只包含接口文档中需要的字段
// 这里只做签名校验和设备校验，业务只返回心跳
//
type QueryCmdRequest struct {
	ViewId       string `json:"ViewId"`
	UID          string `json:"UID"`
	UKey         string `json:"UKey"`
	SN           string `json:"SN"`
	TamperAlarm  string `json:"TamperAlarm"`
	DoorMagnetic string `json:"DoorMagnetic"`
	Timestamp    string `json:"Timestamp"`
	Sign         string `json:"Sign"`
}

func checkSign(req QRRequest, secretKey string) string {
	// 按参数顺序拼接
	var sb strings.Builder
	sb.WriteString("CodeVal=" + req.CodeVal)
	sb.WriteString("CodeType=" + req.CodeType)
	sb.WriteString("BrushTime=" + req.BrushTime)
	sb.WriteString("ViewId=" + req.ViewId)
	sb.WriteString("UID=" + req.UID)
	sb.WriteString("UKey=" + req.UKey)
	sb.WriteString("SN=" + req.SN)
	sb.WriteString("IsOnline=" + req.IsOnline)
	sb.WriteString("Property=" + req.Property)
	sb.WriteString("Timestamp=" + req.Timestamp)
	if secretKey != "" {
		sb.WriteString("SecretKey=" + secretKey)
	}
	h := md5.New()
	h.Write([]byte(sb.String()))
	return hex.EncodeToString(h.Sum(nil))
}

func CheckCodeHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req QRRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("请求解析失败: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// 签名校验
		if req.Sign != "" {
			serverSign := checkSign(req, cfg.SecretKey)
			if !strings.EqualFold(serverSign, req.Sign) {
				log.Printf("签名校验失败，设备UID: %s，期望: %s，实际: %s", req.UID, serverSign, req.Sign)
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("签名校验失败"))
				return
			}
		}
		// 设备识别
		var deviceName string
		for _, d := range cfg.Devices {
			if d.ID == req.UID {
				deviceName = d.Name
				break
			}
		}
		if deviceName == "" {
			log.Printf("未知设备ID(UID): %s", req.UID)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("未知设备ID"))
			return
		}
		log.Printf("收到设备[%s](%s)的核销请求，CodeVal: %s, CodeType: %s, SN: %s, IsOnline: %s, Timestamp: %s", deviceName, req.UID, req.CodeVal, req.CodeType, req.SN, req.IsOnline, req.Timestamp)
		// 组装转发参数
		verifyReq := map[string]interface{}{
			"Verification": 3,
			"McShopID":     cfg.McShopID,
			"IsAllSuccess": true,
			"UseInfoList": []map[string]interface{}{
				{
					"VCode":     req.CodeVal,
					// "CodeType":  req.CodeType,
					// "BrushTime": req.BrushTime,
					// "ViewId":    req.ViewId,
					// "UID":       req.UID,
					// "UKey":      req.UKey,
					// "SN":        req.SN,
					// "IsOnline":  req.IsOnline,
					// "Property":  req.Property,
					// "Timestamp": req.Timestamp,
					// "Sign":      req.Sign,
				},
			},
		}
		verifyBody, _ := json.Marshal(verifyReq)
		resp, err := http.Post(cfg.VerifyURL, "application/json", bytes.NewReader(verifyBody))
		if err != nil {
			log.Printf("核销接口请求失败: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Status":0}`))
			return
		}
		defer resp.Body.Close()
		var verifyResp VerifyResponse
		if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
			log.Printf("核销接口响应解析失败: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Status":0}`))
			return
		}
		if verifyResp.Code == 1 {
			log.Printf("设备[%s](%s)核销成功", deviceName, req.UID)
			w.Header().Set("Content-Type", "application/json")
			// 构造返回体
			respMap := map[string]interface{}{
				"Status": 1,
			}
			// 透传业务系统返回的可选字段
			if verifyResp.Message != "" {
				respMap["StatusDesc"] = verifyResp.Message
			}
			// 支持Relay1Time、BeepType、BeepTime等字段
			var extra map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&extra); err == nil {
				for _, k := range []string{"Relay1Time", "BeepType", "BeepTime"} {
					if v, ok := extra[k]; ok {
						respMap[k] = v
					}
				}
			}
			respBytes, _ := json.Marshal(respMap)
			w.WriteHeader(http.StatusOK)
			w.Write(respBytes)
		} else {
			log.Printf("设备[%s](%s)核销失败: %s", deviceName, req.UID, verifyResp.Message)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Status":0}`))
		}
	}
}

func QueryCmdHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req QueryCmdRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[QueryCmd] 请求解析失败: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"CmdID":"","CmdCode":0,"CmdParams":{}}`))
			return
		}
		// 签名校验
		if req.Sign != "" {
			var sb strings.Builder
			sb.WriteString("ViewId=" + req.ViewId)
			sb.WriteString("UID=" + req.UID)
			sb.WriteString("UKey=" + req.UKey)
			sb.WriteString("SN=" + req.SN)
			sb.WriteString("TamperAlarm=" + req.TamperAlarm)
			sb.WriteString("DoorMagnetic=" + req.DoorMagnetic)
			sb.WriteString("Timestamp=" + req.Timestamp)
			if cfg.SecretKey != "" {
				sb.WriteString("SecretKey=" + cfg.SecretKey)
			}
			h := md5.New()
			h.Write([]byte(sb.String()))
			serverSign := hex.EncodeToString(h.Sum(nil))
			if !strings.EqualFold(serverSign, req.Sign) {
				log.Printf("[QueryCmd] 签名校验失败，设备UID: %s，期望: %s，实际: %s", req.UID, serverSign, req.Sign)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"CmdID":"","CmdCode":0,"CmdParams":{}}`))
				return
			}
		}
		// 设备识别
		var deviceName string
		for _, d := range cfg.Devices {
			if d.ID == req.UID {
				deviceName = d.Name
				break
			}
		}
		if deviceName == "" {
			log.Printf("[QueryCmd] 未知设备ID(UID): %s", req.UID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"CmdID":"","CmdCode":0,"CmdParams":{}}`))
			return
		}
		// 返回心跳命令，CmdID用当前时间戳
		t := time.Now().Format("20060102150405")
		resp := map[string]interface{}{
			"CmdID":    t,
			"CmdCode":  0,
			"CmdParams": map[string]interface{}{ "DateTime": time.Now().Format("2006-01-02 15:04:05") },
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
