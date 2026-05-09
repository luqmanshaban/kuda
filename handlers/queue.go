package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/utils"
)

type QueueHandler struct {
	Repo *repository.QueueRepo
}

func (h *QueueHandler) CreateQH(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name string `json:"name"`
		UserID int 	`json:"user_id"`
		WebhookUrl string `json:"webhook_url"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	if payload.Name == "" || payload.UserID == 0 || payload.WebhookUrl == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "(name, user_id, webhook_url) are required"})
		return
	}

	q, err := h.Repo.CreateQueue(payload.Name, payload.UserID, payload.WebhookUrl)
	if err != nil {
		slog.Error("queue creation failed", "component", "repository", "op", "create_queue", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to create queue"})
		return
	}

	utils.WriteJson(w, http.StatusCreated, q)
}