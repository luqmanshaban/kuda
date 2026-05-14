package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/luqmanshaban/kuda/internal/store"
)

type QueueHandler struct {
	Store *store.QueueStore
}

func (h *QueueHandler) CreateQH(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		Name       string `json:"name"`
		WebhookUrl string `json:"webhook_url"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	if payload.Name == "" || payload.WebhookUrl == "" {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "(name, webhook_url) are required"})
		return
	}

	q, err := h.Store.CreateQueue(payload.Name, payload.WebhookUrl)
	if err != nil {
		slog.Error("queue creation failed", "component", "repository", "op", "create_queue", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to create queue"})
		return
	}

	WriteJson(w, http.StatusCreated, q)
}

func (h *QueueHandler) GetSingleQH(w http.ResponseWriter, r *http.Request) {
	queueName := r.PathValue("name")
	if queueName == "" {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message":"queue name not provided in path"})
		return
	}

	q, err := h.Store.GetQueue(queueName)
	if err != nil {
		slog.Error("queue fetching failed", "component", "repository", "op", "fetch_queue", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to fetch queue"})
		return
	}
	WriteJson(w, http.StatusOK, q)
}

func (h *QueueHandler) GetQH(w http.ResponseWriter, r *http.Request) {
	

	qs, err := h.Store.GetQueues()
	if err != nil {
		slog.Error("queues fetching failed", "component", "repository", "op", "fetch_queue", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to fetch queues"})
		return
	}
	WriteJson(w, http.StatusOK, qs)
}

func (h *QueueHandler) UpdateQH(w http.ResponseWriter, r *http.Request) {


	queueIDstr := r.PathValue("queue_id")

	id, err := strconv.Atoi(queueIDstr)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid id"})
		return
	}

	var payload struct {
		URL string `json:"url"`
	}

	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}
	
	if payload.URL == "" {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "url required"})
		return
	}
	q, err := h.Store.UpdateQueue(payload.URL, id)
	if err != nil {
		slog.Error("queues fetching failed", "component", "repository", "op", "fetch_queue", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to fetch queues"})
		return
	}
	WriteJson(w, http.StatusOK, q)
}
