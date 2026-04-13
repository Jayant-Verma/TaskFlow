package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"taskflow-api/internal/models"
	"taskflow-api/internal/utils"
)

type TaskHandler struct {
	DB *sql.DB
}

type TaskCreateInput struct {
	Title       string  `json:"title" example:"Design Database Schema"`
	Description string  `json:"description" example:"Create the ERD for the new feature"`
	AssigneeID  *string `json:"assignee_id,omitempty" example:"uuid-of-user"`
	DueDate     *string `json:"due_date,omitempty" example:"2025-12-31"`
}

type TaskUpdateInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" enums:"todo,in_progress,done"`
	Priority    *string `json:"priority,omitempty" enums:"low,medium,high"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
}

// List godoc
// @Summary      List tasks
// @Description  Lists tasks for a specific project with optional pagination and filtering.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Param        status query string false "Filter by status" Enums(todo, in_progress, done)
// @Param        assignee query string false "Filter by Assignee UUID"
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(50)
// @Success      200  {array}   models.Task
// @Router       /projects/{id}/tasks [get]
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	query := "SELECT id, title, description, status, priority, project_id, assignee_id, due_date FROM tasks WHERE project_id = $1"
	args := []any{projectID}
	argID := 2

	if s := r.URL.Query().Get("status"); s != "" {
		query += fmt.Sprintf(" AND status = $%d", argID)
		args = append(args, s)
		argID++
	}
	if as := r.URL.Query().Get("assignee"); as != "" {
		query += fmt.Sprintf(" AND assignee_id = $%d", argID)
		args = append(args, as)
		argID++
	}

	limit := 50
	offset := 0
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		limit = l
	}
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		offset = (p - 1) * limit
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	rows, err := h.DB.QueryContext(r.Context(), query, args...)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error", nil)
		return
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var t models.Task
		rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.ProjectID, &t.AssigneeID, &t.DueDate)
		tasks = append(tasks, t)
	}
	utils.WriteJSON(w, http.StatusOK, tasks)
}

// Create godoc
// @Summary      Create a task
// @Description  Creates a new task within a specific project.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Param        request body TaskCreateInput true "Task details"
// @Success      201  {object}  models.Task
// @Failure      400  {object}  map[string]any "Validation failed"
// @Router       /projects/{id}/tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	claims := r.Context().Value(models.UserContextKey).(*models.JWTClaims)

	var input models.Task
	json.NewDecoder(r.Body).Decode(&input)

	if input.Title == "" {
		utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"title": "is required"})
		return
	}

	err := h.DB.QueryRowContext(r.Context(),
		`INSERT INTO tasks (title, description, project_id, creator_id, assignee_id, due_date) 
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		input.Title, input.Description, projectID, claims.UserID, input.AssigneeID, input.DueDate).Scan(&input.ID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error", nil)
		return
	}
	input.ProjectID = projectID
	utils.WriteJSON(w, http.StatusCreated, input)
}

// Update godoc
// @Summary      Update a task
// @Description  Updates a task dynamically based on provided JSON fields.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task UUID"
// @Param        request body TaskUpdateInput true "Fields to update"
// @Success      200  {object}  map[string]string "Status updated"
// @Failure      400  {object}  map[string]any "Invalid fields"
// @Router       /tasks/{id} [patch]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input map[string]any
	json.NewDecoder(r.Body).Decode(&input)

	if len(input) == 0 {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "no changes provided"})
		return
	}

	query := "UPDATE tasks SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}
	argID := 1

	for k, v := range input {
		query += fmt.Sprintf(", %s = $%d", k, argID)
		args = append(args, v)
		argID++
	}
	query += fmt.Sprintf(" WHERE id = $%d", argID)
	args = append(args, id)

	_, err := h.DB.ExecContext(r.Context(), query, args...)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid update fields", nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// Delete godoc
// @Summary      Delete a task
// @Description  Deletes a task. Only the task creator or the project owner can do this.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task UUID"
// @Success      204  "No Content"
// @Failure      403  {object}  map[string]any "Unauthorized action or not found"
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	claims := r.Context().Value(models.UserContextKey).(*models.JWTClaims)

	res, err := h.DB.ExecContext(r.Context(), `
		DELETE FROM tasks t
		USING projects p
		WHERE t.id = $1 AND t.project_id = p.id 
		AND (t.creator_id = $2 OR p.owner_id = $2)
	`, id, claims.UserID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error", nil)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		utils.WriteError(w, http.StatusForbidden, "unauthorized action or not found", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
