package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/response"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/auth"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	authService    *auth.AuthService
	discordService *auth.DiscordService
	googleService  *auth.GoogleService
}

func NewAuthHandler(
	authService *auth.AuthService,
	discordService *auth.DiscordService,
	googleService *auth.GoogleService,
) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		discordService: discordService,
		googleService:  googleService,
	}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/init-register", h.InitRegister)
	r.Post("/register", h.Register)
	r.Post("/resend-email", h.ResendEmail)
	r.Post("/stop-register-session", h.StopRegisterSession)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Post("/refresh", h.Refresh)

	r.Get("/discord/login", h.DiscordLogin)
	r.Get("/discord/callback", h.DiscordCallback)

	r.Get("/google/login", h.GoogleLogin)
	r.Get("/google/callback", h.GoogleCallback)

	return r
}

// InitRegister godoc
// @Summary      Initialize user registration
// @Description  Send email confirmation code and start registration session
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body domain.RegisterDTO true "User registration details"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ValidationErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /auth/init-register [post]
func (h *AuthHandler) InitRegister(w http.ResponseWriter, r *http.Request) {
	var dto domain.RegisterDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if errs := dto.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	if err := h.authService.InitRegister(w, &dto); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "can not resend") || strings.Contains(err.Error(), "maximum verification attempts") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"sent": true})
}

// Register godoc
// @Summary      Complete user registration
// @Description  Verify email code and create user account
// @Tags         auth
// @Produce      json
// @Param        code query int true "Email verification code"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	codeStr := r.URL.Query().Get("code")
	if codeStr == "" {
		response.Error(w, http.StatusBadRequest, "Validation failed (numeric string is expected)")
		return
	}

	code, err := strconv.Atoi(codeStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed (numeric string is expected)")
		return
	}

	res, err := h.authService.Register(w, r, code)
	if err != nil {
		if strings.Contains(err.Error(), "register token") || strings.Contains(err.Error(), "not found") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid code") {
			response.Error(w, http.StatusUnauthorized, err.Error())
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// ResendEmail godoc
// @Summary      Resend verification email
// @Description  Resend 6-digit confirmation code for active registration session
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /auth/resend-email [post]
func (h *AuthHandler) ResendEmail(w http.ResponseWriter, r *http.Request) {
	if err := h.authService.ResendEmail(w, r); err != nil {
		if strings.Contains(err.Error(), "register token") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "can not resend") || strings.Contains(err.Error(), "maximum verification attempts") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"resent": true})
}

// StopRegisterSession godoc
// @Summary      Cancel registration session
// @Description  Clear registration session cookies and temporary data
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Router       /auth/stop-register-session [post]
func (h *AuthHandler) StopRegisterSession(w http.ResponseWriter, r *http.Request) {
	h.authService.StopRegisterSession(w)
	response.JSON(w, http.StatusOK, map[string]any{"sessionStopped": true})
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body domain.LoginDTO true "Login credentials"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ValidationErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var dto domain.LoginDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if errs := dto.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	res, err := h.authService.Login(w, &dto)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found!")
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// Logout godoc
// @Summary      User logout
// @Description  Clear authentication cookies and session
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authService.Logout(w)
	response.JSON(w, http.StatusOK, map[string]any{"loggedOut": true})
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Exchange refresh token cookie for new access token
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	res, err := h.authService.Refresh(w, r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// DiscordLogin godoc
// @Summary      Discord OAuth redirect
// @Description  Redirect to Discord OAuth2 authorization URL
// @Tags         auth
// @Success      302 {string} string "Redirect to Discord"
// @Router       /auth/discord/login [get]
func (h *AuthHandler) DiscordLogin(w http.ResponseWriter, r *http.Request) {
	authURL := h.discordService.GetAuthURL()
	http.Redirect(w, r, authURL, http.StatusFound)
}

// DiscordCallback godoc
// @Summary      Discord OAuth callback
// @Description  Handle Discord OAuth2 code and sign user in
// @Tags         auth
// @Produce      json
// @Param        code query string true "Discord authorization code"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Router       /auth/discord/callback [get]
func (h *AuthHandler) DiscordCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		response.Error(w, http.StatusBadRequest, "Code parameter is missing")
		return
	}

	user, refreshToken, expiresIn, err := h.discordService.HandleCallback(r.Context(), code)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	h.authService.SetCookie(w, "refreshToken", refreshToken, time.Now().Add(expiresIn))
	res, err := h.authService.SetAuthResponse(w, user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// GoogleLogin godoc
// @Summary      Google OAuth redirect
// @Description  Redirect to Google OAuth2 authorization URL
// @Tags         auth
// @Success      302 {string} string "Redirect to Google"
// @Router       /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	authURL := h.googleService.GetAuthURL()
	http.Redirect(w, r, authURL, http.StatusFound)
}

// GoogleCallback godoc
// @Summary      Google OAuth callback
// @Description  Handle Google OAuth2 code and sign user in
// @Tags         auth
// @Produce      json
// @Param        code query string true "Google authorization code"
// @Success      200 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Router       /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		response.Error(w, http.StatusBadRequest, "Code parameter is missing")
		return
	}

	user, refreshToken, expiresIn, err := h.googleService.HandleCallback(r.Context(), code)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	h.authService.SetCookie(w, "refreshToken", refreshToken, time.Now().Add(expiresIn))
	res, err := h.authService.SetAuthResponse(w, user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, res)
}
