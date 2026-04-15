package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"taskflow-api/internal/models"
	"taskflow-api/internal/utils"
)

type TaskHandler struct {
	DB *sql.DB
}

type TaskCreateInput struct {
	Title       string     `json:"title" example:"Design Database Schema"`
	Description string     `json:"description" example:"Create the ERD for the new feature"`
	Priority    string     `json:"priority" example:"medium" enums:"low,medium,high"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty" example:"uuid-of-user"`
	DueDate     *time.Time `json:"due_date,omitempty" example:"2025-12-31"`
}

type TaskUpdateInput struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Status      *string    `json:"status,omitempty" enums:"todo,in_progress,done"`
	Priority    *string    `json:"priority,omitempty" enums:"low,medium,high"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

var validTaskStatuses = map[string]bool{
	"todo":        true,
	"in_progress": true,
	"done":        true,
}

var validTaskPriorities = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
}

func isValidStatus(value string) bool {
	return validTaskStatuses[value]
}

func isValidPriority(value string) bool {
	return validTaskPriorities[value]
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
		if !isValidStatus(s) {
			utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"status": "invalid value"})
			return
		}
		query += fmt.Sprintf(" AND status = $%d", argID)
		args = append(args, s)
		argID++
	}
	if as := r.URL.Query().Get("assignee"); as != "" {
		if _, err := uuid.Parse(as); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"assignee": "must be a valid UUID"})
			return
		}
		query += fmt.Sprintf(" AND assignee_id = $%d", argID)
		args = append(args, as)
		argID++
	}

	limit := 50
	offset := 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		l, err := strconv.Atoi(lStr)
		if err != nil || l <= 0 {
			utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"limit": "must be a positive integer"})
			return
		}
		limit = l
	}
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		p, err := strconv.Atoi(pStr)
		if err != nil || p <= 0 {
			utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"page": "must be a positive integer"})
			return
		}
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

	var input TaskCreateInput
	if !utils.DecodeJSON(w, r, &input) {
		return
	}

	if input.Title == "" {
		utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"title": "is required"})
		return
	}

	task := models.Task{
		Title:       input.Title,
		Description: input.Description,
		ProjectID:   uuid.MustParse(projectID),
		AssigneeID:  input.AssigneeID,
		DueDate:     input.DueDate,
	}

	err := h.DB.QueryRowContext(r.Context(),
		`INSERT INTO tasks (title, description, project_id, creator_id, assignee_id, due_date) 
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		task.Title, task.Description, projectID, claims.UserID, task.AssigneeID, task.DueDate).Scan(&task.ID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error", nil)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, task)
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
	var input TaskUpdateInput
	if !utils.DecodeJSON(w, r, &input) {
		return
	}

	query := "UPDATE tasks SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}
	argID := 1

	if input.Title != nil {
		if *input.Title == "" {
			utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"title": "cannot be empty"})
			return
		}
		query += fmt.Sprintf(", title = $%d", argID)
		args = append(args, *input.Title)
		argID++
	}
	if input.Description != nil {
		query += fmt.Sprintf(", description = $%d", argID)
		args = append(args, *input.Description)
		argID++
	}
	if input.Status != nil {
		if !isValidStatus(*input.Status) {
			utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"status": "invalid value"})
			return
		}
		query += fmt.Sprintf(", status = $%d", argID)
		args = append(args, *input.Status)
		argID++
	}
	if input.Priority != nil {
		if !isValidPriority(*input.Priority) {
			utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"priority": "invalid value"})
			return
		}
		query += fmt.Sprintf(", priority = $%d", argID)
		args = append(args, *input.Priority)
		argID++
	}
	if input.AssigneeID != nil {
		query += fmt.Sprintf(", assignee_id = $%d", argID)
		args = append(args, input.AssigneeID)
		argID++
	}
	if input.DueDate != nil {
		query += fmt.Sprintf(", due_date = $%d", argID)
		args = append(args, input.DueDate)
		argID++
	}

	if len(args) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"body": "no fields to update"})
		return
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
