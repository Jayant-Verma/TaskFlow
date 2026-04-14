package handlers

import (
	"database/sql"
	"net/http"

	"taskflow-api/internal/models"
	"taskflow-api/internal/utils"
)

type UserHandler struct {
	DB *sql.DB
}

// List godoc
// @Summary      List users
// @Description  Retrieves all users.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.User
// @Failure      401  {object}  map[string]any "Unauthenticated"
// @Router       /users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.QueryContext(r.Context(), `
		SELECT DISTINCT u.id, u.name, u.email, u.created_at
		FROM users u
	`)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "DB Error", nil)
		return
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
		users = append(users, u)
	}
	utils.WriteJSON(w, http.StatusOK, users)
}
