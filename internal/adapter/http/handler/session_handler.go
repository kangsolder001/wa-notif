package handler

import (
	"encoding/json"
	"net/http"
	"wa-notif/internal/usecase"
	"wa-notif/internal/usecase/dto"
)

type SessionHandler struct {
	sessionUC *usecase.SessionUseCase
}

func NewSessionHandler(sessionUC *usecase.SessionUseCase) *SessionHandler {
	return &SessionHandler{
		sessionUC: sessionUC,
	}
}

func (h *SessionHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := h.sessionUC.GetStatus()

	_ = json.NewEncoder(w).Encode(dto.APIResponse{
		Success: true,
		Data:    status,
	})
}

func (h *SessionHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	qrRes, err := h.sessionUC.GetQRCode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "QR_GENERATION_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	_ = json.NewEncoder(w).Encode(dto.APIResponse{
		Success: true,
		Data:    qrRes,
	})
}

func (h *SessionHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := h.sessionUC.Logout(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "LOGOUT_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	_ = json.NewEncoder(w).Encode(dto.APIResponse{
		Success: true,
		Data: map[string]string{
			"status":  "DISCONNECTED",
			"message": "Device successfully unlinked / logged out",
		},
	})
}
