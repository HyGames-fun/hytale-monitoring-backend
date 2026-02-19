package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/middleware"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/response"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userRepo *postgres.UserRepo
}

func NewUserHandler(userRepo *postgres.UserRepo) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetCurrentUser)
	r.Put("/avatar", h.SetAvatar)
	r.Delete("/avatar", h.DeleteAvatar)
	r.Get("/avatar", h.GetSelfAvatar)
	r.Get("/avatar/{id}", h.GetAvatarByID)

	return r
}

// GetCurrentUser godoc
// @Summary      Get current authenticated user
// @Description  Get profile of currently logged-in user
// @Tags         user
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /user [get]
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	res := map[string]any{
		"name":   user.Name,
		"email":  user.Email,
		"avatar": user.Avatar,
	}
	response.JSON(w, http.StatusOK, res)
}

type SetAvatarDTO struct {
	AvatarURL string `json:"avatarUrl"`
}

// SetAvatar godoc
// @Summary      Set user avatar
// @Description  Update user profile avatar by URL or multipart upload
// @Tags         user
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Accept       json
// @Produce      json
// @Param        avatar formData file false "Avatar image upload"
// @Param        input body SetAvatarDTO false "Avatar URL payload"
// @Success      200 {object} response.SuccessResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /user/avatar [put]
func (h *UserHandler) SetAvatar(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var avatarURL string
	if err := r.ParseMultipartForm(5 << 20); err == nil {
		avatarURL = fmt.Sprintf("https://s3.storage.hytale-monitoring.fun/avatars/%d.webp", user.ID)
	} else {
		var dto SetAvatarDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err == nil {
			avatarURL = dto.AvatarURL
		} else {
			avatarURL = fmt.Sprintf("https://s3.storage.hytale-monitoring.fun/avatars/%d.webp", user.ID)
		}
	}

	if err := h.userRepo.SetAvatar(user.ID, &avatarURL); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"avatar": avatarURL})
}

// DeleteAvatar godoc
// @Summary      Delete user avatar
// @Description  Remove avatar from user profile
// @Tags         user
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /user/avatar [delete]
func (h *UserHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.userRepo.SetAvatar(user.ID, nil); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"avatar": nil})
}

// GetSelfAvatar godoc
// @Summary      Get current user avatar
// @Description  Get avatar URL of current authenticated user
// @Tags         user
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /user/avatar [get]
func (h *UserHandler) GetSelfAvatar(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	response.JSON(w, http.StatusOK, user.Avatar)
}

// GetAvatarByID godoc
// @Summary      Get user avatar by ID
// @Description  Get avatar URL of any user by their user ID
// @Tags         user
// @Produce      json
// @Param        id path int true "User ID"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Router       /user/avatar/{id} [get]
func (h *UserHandler) GetAvatarByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user id")
		return
	}

	targetUser, err := h.userRepo.FindUserByID(userID)
	if err != nil || targetUser == nil {
		response.Error(w, http.StatusBadRequest, "User not found")
		return
	}

	response.JSON(w, http.StatusOK, targetUser.Avatar)
}
