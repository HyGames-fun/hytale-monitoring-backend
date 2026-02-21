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
	r.Get("/docs", ScalarHandler)
	r.Get("/reference", ScalarHandler)
	r.Get("/reference/*", ScalarHandler)
	r.Get("/openapi.json", OpenAPIHandler)
	r.Get("/swagger.json", OpenAPIHandler)
	r.Get("/docs/openapi.json", OpenAPIHandler)

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
	})

	return r
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Check service health status
// @Tags         health
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Router       / [get]
func (h *CatalogHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// CreateServer godoc
// @Summary      Create a new server
// @Description  Register a new Hytale server in the monitoring catalog
// @Tags         server
// @Accept       json
// @Produce      json
// @Param        input body domain.CreateServerDTO true "Server creation payload"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ValidationErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /server/create [post]
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

// LikeServer godoc
// @Summary      Like or unlike a server
// @Description  Vote for a server with Turnstile verification or user session
// @Tags         server
// @Produce      json
// @Param        id query int true "Server ID"
// @Param        token query string false "Cloudflare Turnstile token for guests"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /server/like [post]
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

// FindPage godoc
// @Summary      Find servers page
// @Description  Retrieve paginated list of servers with filters and sorting
// @Tags         server
// @Accept       json
// @Produce      json
// @Param        input body domain.FindPageDTO true "Pagination and filter criteria"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ValidationErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /server [post]
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

// FindQuantity godoc
// @Summary      Get total servers count
// @Description  Get the total count of registered servers
// @Tags         server
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /server/quantity [get]
func (h *CatalogHandler) FindQuantity(w http.ResponseWriter, r *http.Request) {
	count, err := h.serverService.FindQuantity()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, count)
}

// GetStatus godoc
// @Summary      Get online status of a server
// @Description  Check online status of a server by ID
// @Tags         server
// @Produce      json
// @Param        id query int true "Server ID"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Router       /server/status [get]
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

// GetStatuses godoc
// @Summary      Get online statuses for multiple servers
// @Description  Batch check online statuses by server IDs
// @Tags         server
// @Produce      json
// @Param        ids query string true "Comma-separated server IDs"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Router       /server/statuses [get]
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

// Search godoc
// @Summary      Search servers
// @Description  Full-text search servers by name, description or tags
// @Tags         server
// @Produce      json
// @Param        query query string false "Search query"
// @Param        page query int false "Page number"
// @Param        limit query int false "Items per page"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Router       /server/search [get]
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

// FindOne godoc
// @Summary      Find server by nameId
// @Description  Get detailed server information by kebab-case nameId
// @Tags         server
// @Produce      json
// @Param        nameId path string true "Server kebab-case nameId"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Router       /server/id/{nameId} [get]
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

// UpdateServer godoc
// @Summary      Update server details
// @Description  Update server metadata by ID
// @Tags         server
// @Accept       json
// @Produce      json
// @Param        id path int true "Server ID"
// @Param        input body domain.UpdateServerDTO true "Update server payload"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ValidationErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /server/id/{id} [patch]
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

// DeleteServer godoc
// @Summary      Delete server
// @Description  Delete server by ID
// @Tags         server
// @Produce      json
// @Param        id path int true "Server ID"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /server/id/{id} [delete]
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

// GetLeaderboard godoc
// @Summary      Get servers leaderboard
// @Description  Get top servers ranked by score (likes and active players)
// @Tags         server
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /server/leaderboard [get]
func (h *CatalogHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboard, err := h.serverService.GetLeaderboard()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, leaderboard)
}
