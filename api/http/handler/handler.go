package handler

import (
	"encoding/json"
	"net/http"
)

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Health 健康检查
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data: map[string]string{
			"status": "healthy",
		},
	})
}

// GetBalance 查询余额
func GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, Response{
			Code:    -1,
			Message: "方法不允许",
		})
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Code:    -1,
			Message: "缺少 address 参数",
		})
		return
	}

	// TODO: 实现查询余额逻辑
	writeJSON(w, http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"address": address,
			"balance": "1.234567",
		},
	})
}

// Transfer 转账
func Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{
			Code:    -1,
			Message: "方法不允许",
		})
		return
	}

	var req struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount string `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{
			Code:    -1,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// TODO: 实现转账逻辑
	writeJSON(w, http.StatusOK, Response{
		Code:    0,
		Message: "转账成功",
		Data: map[string]string{
			"tx_hash": "0x123...",
		},
	})
}

// GetTransactions 查询交易记录
func GetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, Response{
			Code:    -1,
			Message: "方法不允许",
		})
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Code:    -1,
			Message: "缺少 address 参数",
		})
		return
	}

	// TODO: 实现查询交易记录逻辑
	writeJSON(w, http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: []map[string]interface{}{
			{
				"tx_hash": "0x123...",
				"from":    "0xabc...",
				"to":      address,
				"amount":  "1.0",
			},
		},
	})
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
