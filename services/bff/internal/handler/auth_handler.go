package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	userv1 "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	user       userv1.UserServiceClient
	jwtManager *auth.JWTManager
	logger     *slog.Logger
}

func NewAuthHandler(user userv1.UserServiceClient, jwtManager *auth.JWTManager, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		user:       user,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	ctx := r.Context()
	resp, err := h.user.Login(ctx, &userv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			writeJSONError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		h.logger.Error("login failed", "error", err, "email", req.Email)
		writeJSONError(w, http.StatusInternalServerError, "login failed")
		return
	}

	token, err := h.jwtManager.Generate(resp.User.Id, resp.User.Email)
	if err != nil {
		h.logger.Error("failed to generate token", "error", err, "user_id", resp.User.Id)
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(LoginResponse{
		Token:     token,
		ExpiresIn: 86400,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	ctx := r.Context()
	resp, err := h.user.CreateUser(ctx, &userv1.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument || status.Code(err) == codes.AlreadyExists {
			writeJSONError(w, http.StatusBadRequest, status.Convert(err).Message())
			return
		}
		h.logger.Error("registration failed", "error", err, "email", req.Email)
		writeJSONError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(RegisterResponse{
		User: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}{
			ID:    resp.User.Id,
			Email: resp.User.Email,
		},
	})
}
