package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ContextKey string

const UserContextKey = ContextKey("user")

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name      string    `json:"name"`
	Email     string    `json:"email" gorm:"unique"`
	Hash      string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Task struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	ProjectID   uuid.UUID  `json:"project_id"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}
