package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockVerifyResponse struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
}

func startMockVerifyServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := mockVerifyResponse{Code: 1, Message: "合法卡"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestCheckCodeHandler_Success(t *testing.T) {
	mockServer := startMockVerifyServer()
	defer mockServer.Close()

	cfg := &Config{
		ListenIP:  "127.0.0.1",
		Port:      "8080",
		VerifyURL: mockServer.URL,
		McShopID:  123456,
		SecretKey: "testkey",
		Devices:   []Device{{ID: "dev001", Name: "测试设备"}},
	}

	// 构造签名
	reqBody := QRRequest{
		CodeVal:   "998678",
		CodeType:  "Q",
		BrushTime: "2025-05-26 10:20:30",
		ViewId:    "D2",
		UID:       "dev001",
		UKey:      "3F698DAC58",
		SN:        "2001000111",
		IsOnline:  "1",
		Property:  "1",
		Timestamp: "1591789801",
	}
	reqBody.Sign = checkSign(reqBody, cfg.SecretKey)

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/CheckCode", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := CheckCodeHandler(cfg)
	handler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("期望状态码200，实际: %d", resp.StatusCode)
	}
	var respMap map[string]interface{}
	if err := json.Unmarshal(body, &respMap); err != nil {
		t.Fatalf("响应体不是合法JSON: %v", err)
	}
	if respMap["Status"] != float64(1) {
		t.Fatalf("期望Status=1，实际: %v", respMap["Status"])
	}
	if respMap["StatusDesc"] != "合法卡" {
		t.Fatalf("期望StatusDesc=合法卡，实际: %v", respMap["StatusDesc"])
	}
}
