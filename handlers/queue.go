package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/structs"
	"github.com/luqmanshaban/kuda/utils"
)

type QueueHandler struct {
	Repo *repository.QueueRepo
}

func (h *QueueHandler) CreateQH(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("user").(structs.User).ID
	var payload struct {
		Name       string `json:"name"`
		WebhookUrl string `json:"webhook_url"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	if payload.Name == "" || payload.WebhookUrl == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "(name, webhook_url) are required"})
		return
	}

	q, err := h.Repo.CreateQueue(payload.Name, id, payload.WebhookUrl)
	if err != nil {
		slog.Error("queue creation failed", "component", "repository", "op", "create_queue", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to create queue"})
		return
	}

	utils.WriteJson(w, http.StatusCreated, q)
}

func (h *QueueHandler) GetQH(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("user").(structs.User).ID

	qs, err := h.Repo.GetQueues(id)
	if err != nil {
		slog.Error("queues fetching failed", "component", "repository", "op", "fetch_queue", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to fetch queues"})
		return
	}
	utils.WriteJson(w, http.StatusOK, qs)
}

func (h *QueueHandler) UpdateQH(w http.ResponseWriter, r *http.Request) {
	uId := r.Context().Value("user").(structs.User).ID

	queueIDstr := r.PathValue("queue_id")

	id, err := strconv.Atoi(queueIDstr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid id"})
		return
	}

	var payload struct {
		URL string `json:"url"`
	}

	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}
	
	if payload.URL == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "url required"})
		return
	}
	q, err := h.Repo.UpdateQueue(payload.URL, id, uId)
	if err != nil {
		slog.Error("queues fetching failed", "component", "repository", "op", "fetch_queue", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to fetch queues"})
		return
	}
	utils.WriteJson(w, http.StatusOK, q)
}
