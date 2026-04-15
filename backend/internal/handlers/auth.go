package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"taskflow-api/internal/models"
	"taskflow-api/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB        *sql.DB
	JWTSecret []byte
}

type RegisterInput struct {
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"securepassword123"`
}

type LoginInput struct {
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"securepassword123"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with a hashed password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterInput true "Registration details"
// @Success      201  {object}  map[string]string "Returns the new user ID"
// @Failure      400  {object}  map[string]any "Validation failed or email exists"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if !utils.DecodeJSON(w, r, &input) {
		return
	}

	fields := make(map[string]string)
	if input.Name == "" {
		fields["name"] = "is required"
	}
	if input.Email == "" {
		fields["email"] = "is required"
	}
	if input.Password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		utils.WriteError(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 12)

	var id string
	err := h.DB.QueryRowContext(r.Context(),
		"INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id",
		input.Name, input.Email, hashed).Scan(&id)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "email may already exist", nil)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// Login godoc
// @Summary      Login a user
// @Description  Authenticates a user and returns a 24-hour JWT access token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body LoginInput true "Login credentials"
// @Success      200  {object}  map[string]string "Returns access_token"
// @Failure      401  {object}  map[string]any "Invalid credentials"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if !utils.DecodeJSON(w, r, &input) {
		return
	}

	fields := make(map[string]string)
	if input.Email == "" {
		fields["email"] = "is required"
	}
	if input.Password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		utils.WriteError(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	var id, hash string
	err := h.DB.QueryRowContext(r.Context(), "SELECT id, password FROM users WHERE email = $1", input.Email).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)) != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	claims := &models.JWTClaims{
		UserID: id,
		Email:  input.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString(h.JWTSecret)

	utils.WriteJSON(w, http.StatusOK, map[string]string{"access_token": t})
}
