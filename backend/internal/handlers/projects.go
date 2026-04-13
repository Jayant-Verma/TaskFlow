package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"taskflow-api/internal/models"
	"taskflow-api/internal/utils"
)

type ProjectHandler struct {
	DB *sql.DB
}

type ProjectInput struct {
	Name        string `json:"name" example:"Website Redesign"`
	Description string `json:"description" example:"Overhaul of the main landing page"`
}

// List godoc
// @Summary      List projects
// @Description  Retrieves all projects the current user owns or has tasks assigned in.
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Project
// @Failure      401  {object}  map[string]any "Unauthenticated"
// @Router       /projects [get]
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(models.UserContextKey).(*models.JWTClaims)

	rows, err := h.DB.QueryContext(r.Context(), `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at 
		FROM projects p 
		LEFT JOIN tasks t ON p.id = t.project_id 
		WHERE p.owner_id = $1 OR t.assignee_id = $1
	`, claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error", nil)
		return
	}
	defer rows.Close()

	projects := []models.Project{}
	for rows.Next() {
		var p models.Project
		rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
		projects = append(projects, p)
	}
	utils.WriteJSON(w, http.StatusOK, projects)
}

// Create godoc
// @Summary      Create a project
// @Description  Creates a new project owned by the current user.
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body ProjectInput true "Project details"
// @Success      201  {object}  models.Project
// @Failure      400  {object}  map[string]any "Validation failed"
// @Failure      401  {object}  map[string]any "Unauthenticated"
// @Router       /projects [post]
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(models.UserContextKey).(*models.JWTClaims)
	var input ProjectInput
	json.NewDecoder(r.Body).Decode(&input)

	if input.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, "validation failed", map[string]string{"name": "is required"})
		return
	}

	var p models.Project
	err := h.DB.QueryRowContext(r.Context(),
		"INSERT INTO projects (name, description, owner_id) VALUES ($1, $2, $3) RETURNING id, name, description, owner_id, created_at",
		input.Name, input.Description, claims.UserID).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error", nil)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, p)
}

// Get godoc
// @Summary      Get a project
// @Description  Retrieves project details and a summary of its tasks.
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Success      200  {object}  map[string]any "Contains 'project' and 'tasks' array"
// @Failure      404  {object}  map[string]any "Not found"
// @Router       /projects/{id} [get]
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var p models.Project
	err := h.DB.QueryRowContext(r.Context(), "SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1", id).
		Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)

	if err == sql.ErrNoRows {
		utils.WriteError(w, http.StatusNotFound, "not found", nil)
		return
	}

	rows, _ := h.DB.QueryContext(r.Context(), "SELECT id, title, status FROM tasks WHERE project_id = $1", id)
	defer rows.Close()

	tasks := []map[string]any{}
	for rows.Next() {
		var tid, ttitle, tstatus string
		rows.Scan(&tid, &ttitle, &tstatus)
		tasks = append(tasks, map[string]any{"id": tid, "title": ttitle, "status": tstatus})
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"project": p, "tasks": tasks})
}

// Update godoc
// @Summary      Update a project
// @Description  Updates project details. Only the owner can do this.
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Param        request body ProjectInput true "Updated fields (optional)"
// @Success      200  {object}  map[string]string "Status updated"
// @Failure      403  {object}  map[string]any "Unauthorized action or not found"
// @Router       /projects/{id} [patch]
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	claims := r.Context().Value(models.UserContextKey).(*models.JWTClaims)

	var input ProjectInput
	json.NewDecoder(r.Body).Decode(&input)

	res, err := h.DB.ExecContext(r.Context(), "UPDATE projects SET name = COALESCE(NULLIF($1, ''), name), description = COALESCE(NULLIF($2, ''), description) WHERE id = $3 AND owner_id = $4", input.Name, input.Description, id, claims.UserID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error", nil)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		utils.WriteError(w, http.StatusForbidden, "unauthorized action or not found", nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// Delete godoc
// @Summary      Delete a project
// @Description  Deletes a project and all associated tasks. Only the owner can do this.
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Success      204  "No Content"
// @Failure      403  {object}  map[string]any "Unauthorized action or not found"
// @Router       /projects/{id} [delete]
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	claims := r.Context().Value(models.UserContextKey).(*models.JWTClaims)

	res, _ := h.DB.ExecContext(r.Context(), "DELETE FROM projects WHERE id = $1 AND owner_id = $2", id, claims.UserID)
	affected, _ := res.RowsAffected()
	if affected == 0 {
		utils.WriteError(w, http.StatusForbidden, "unauthorized action or not found", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Stats godoc
// @Summary      Project Task Stats
// @Description  Returns task counts grouped by status.
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Success      200  {object}  map[string]any "Contains status_counts map"
// @Router       /projects/{id}/stats [get]
func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	rows, _ := h.DB.QueryContext(r.Context(), "SELECT status, COUNT(*) FROM tasks WHERE project_id = $1 GROUP BY status", id)
	defer rows.Close()

	statusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		statusCounts[status] = count
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"status_counts": statusCounts})
}
