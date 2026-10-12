package handler

import (
	"encoding/json"
	"net/http"
	"wa-notif/internal/usecase"
	"wa-notif/internal/usecase/dto"
)

type MessageHandler struct {
	sendMsgUC *usecase.SendMessageUseCase
}

func NewMessageHandler(sendMsgUC *usecase.SendMessageUseCase) *MessageHandler {
	return &MessageHandler{
		sendMsgUC: sendMsgUC,
	}
}

func (h *MessageHandler) SendText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dto.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Malformed JSON payload: " + err.Error(),
			},
		})
		return
	}

	res, err := h.sendMsgUC.SendTextMessage(r.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "SEND_FAILED"
		if err == usecase.ErrNotConnected {
			status = http.StatusServiceUnavailable
			code = "WA_NOT_CONNECTED"
		}

		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    code,
				Message: err.Error(),
			},
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.APIResponse{
		Success: true,
		Data:    res,
	})
}

func (h *MessageHandler) SendMedia(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dto.SendMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Malformed JSON payload: " + err.Error(),
			},
		})
		return
	}

	res, err := h.sendMsgUC.SendMediaMessage(r.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "SEND_FAILED"
		if err == usecase.ErrNotConnected {
			status = http.StatusServiceUnavailable
			code = "WA_NOT_CONNECTED"
		}

		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    code,
				Message: err.Error(),
			},
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.APIResponse{
		Success: true,
		Data:    res,
	})
}
