package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/middleware"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/response"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/catalog"
	"github.com/go-chi/chi/v5"
)

type CatalogHandler struct {
	serverService       *catalog.ServerService
	verificationService *catalog.VerificationService
}

func NewCatalogHandler(serverService *catalog.ServerService, verificationService *catalog.VerificationService) *CatalogHandler {
	return &CatalogHandler{
		serverService:       serverService,
		verificationService: verificationService,
	}
}

func (h *CatalogHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.HealthCheck)

	r.Route("/server", func(sr chi.Router) {
		sr.Post("/create", h.CreateServer)
		sr.Post("/like", h.LikeServer)
		sr.Post("/", h.FindPage)
		sr.Get("/quantity", h.FindQuantity)
		sr.Get("/statuses", h.GetStatuses)
		sr.Get("/status", h.GetStatus)
		sr.Get("/search", h.Search)
		sr.Get("/leaderboard", h.GetLeaderboard)
		sr.Get("/id/{nameId}", h.FindOne)
		sr.Patch("/id/{id}", h.UpdateServer)
		sr.Delete("/id/{id}", h.DeleteServer)

		sr.Post("/{id}/verify/init", h.InitVerification)
		sr.Post("/{id}/verify/confirm", h.ConfirmVerification)
	})

	return r
}

func (h *CatalogHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *CatalogHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateServerDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if errs := dto.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	server, err := h.serverService.Create(&dto)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			response.Error(w, http.StatusConflict, "Server with domain, ip, or nameId already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, server)
}

func (h *CatalogHandler) LikeServer(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		response.Error(w, http.StatusBadRequest, "Validation failed (numeric string is expected)")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed (numeric string is expected)")
		return
	}

	token := r.URL.Query().Get("token")
	user := middleware.GetUserFromContext(r.Context())
	ip := middleware.GetClientIP(r)

	isLiked, err := h.serverService.Like(id, user, token, ip)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "turnstile") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"liked": isLiked})
}

func (h *CatalogHandler) FindPage(w http.ResponseWriter, r *http.Request) {
	var dto domain.FindPageDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if errs := dto.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	ip := middleware.GetClientIP(r)

	servers, err := h.serverService.FindPage(&dto, user, ip)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, servers)
}

func (h *CatalogHandler) FindQuantity(w http.ResponseWriter, r *http.Request) {
	count, err := h.serverService.FindQuantity()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, count)
}

func (h *CatalogHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		response.Error(w, http.StatusBadRequest, "Validation failed (numeric string is expected)")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed (numeric string is expected)")
		return
	}

	status, err := h.serverService.GetOnlineStatus(id)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, status)
}

func (h *CatalogHandler) GetStatuses(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query()["ids"]
	if len(idsParam) == 0 {
		idsParam = strings.Split(r.URL.Query().Get("ids"), ",")
	}

	var ids []int
	for _, s := range idsParam {
		for _, part := range strings.Split(s, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				if id, err := strconv.Atoi(trimmed); err == nil {
					ids = append(ids, id)
				}
			}
		}
	}

	if len(ids) == 0 {
		response.Error(w, http.StatusBadRequest, "Validation failed (parsable array expected)")
		return
	}

	statuses, err := h.serverService.GetOnlineStatuses(ids)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, statuses)
}

func (h *CatalogHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("query")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if limit <= 0 {
		limit = 10
	}

	results, err := h.serverService.Search(q, page, limit)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, results)
}

func (h *CatalogHandler) FindOne(w http.ResponseWriter, r *http.Request) {
	nameID := chi.URLParam(r, "nameId")
	user := middleware.GetUserFromContext(r.Context())
	ip := middleware.GetClientIP(r)

	server, err := h.serverService.FindOne(nameID, user, ip)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Server not found")
		return
	}

	response.JSON(w, http.StatusOK, server)
}

func (h *CatalogHandler) UpdateServer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid id")
		return
	}

	var dto domain.UpdateServerDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if errs := dto.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	server, err := h.serverService.Update(id, &dto)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Server not found")
		return
	}

	response.JSON(w, http.StatusOK, server)
}

func (h *CatalogHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid id")
		return
	}

	server, err := h.serverService.Remove(id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Server not found")
		return
	}

	response.JSON(w, http.StatusOK, server)
}

func (h *CatalogHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboard, err := h.serverService.GetLeaderboard()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, leaderboard)
}

func (h *CatalogHandler) InitVerification(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	serverID, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid server id")
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	userID := 1
	if user != nil {
		userID = user.ID
	}

	resp, err := h.verificationService.InitVerification(serverID, userID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *CatalogHandler) ConfirmVerification(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	serverID, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid server id")
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	userID := 1
	if user != nil {
		userID = user.ID
	}

	ok, msg, err := h.verificationService.ConfirmDNSVerification(serverID, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		response.Error(w, http.StatusBadRequest, msg)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"verified": true, "message": msg})
}
