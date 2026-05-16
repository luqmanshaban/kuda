// dead letter jobs
package handlers

import (
	"net/http"

	"github.com/luqmanshaban/kuda/internal/store"
)

type DeadLetterHandler struct {
	Store *store.DeadLetterJobStore
}

func (h *DeadLetterHandler) GetDeadJobs(w http.ResponseWriter, r *http.Request) {
	js, err := h.Store.GetDeadJobs();
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, map[string]any{"message": "failed to fetch dead jobs", "error": err})
		return
	}

	WriteJson(w, http.StatusOK, js)
}
