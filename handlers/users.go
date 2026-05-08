package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/utils"
)

type UserHandler struct {
	Repo *repository.UserRepository
}

func (h *UserHandler) CreateUH(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "email not provided"})
		return
	}

	pass, err := utils.HashPasswd(payload.Password)
	payload.Password = pass

	if err != nil {
		fmt.Println(err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create user"})
		return
	}
	u, err := h.Repo.CreateUser(payload.Email, payload.Password)
	if err != nil {
		fmt.Println(err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create user"})
		return
	}

	utils.WriteJson(w, http.StatusCreated, u)
}

func (h *UserHandler) GetUserJH(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("user_id")
	if userId == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "job id not provided"})
		return
	}

	id, err := strconv.Atoi(userId)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "job id format is invalid"})
		return
	}

	// CHECK if query filters are passed
	statuses := []string{"pending", "running", "completed", "failed"}
	param := r.URL.Query().Get("status")
	for _, status := range statuses {
		if param == status {
			j, err := h.Repo.GetFilteredUserJobs(id, status)
			fmt.Println("jobs length: ", len(j))
			if err != nil {
				fmt.Println(err)
				utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch for job"})
				return
			}
			utils.WriteJson(w, http.StatusOK, j)
			return
		}
	}

	j, err := h.Repo.GetUserJobs(id)
	if err != nil {
		fmt.Println(err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch for job"})
		return
	}

	utils.WriteJson(w, http.StatusOK, j)
}
