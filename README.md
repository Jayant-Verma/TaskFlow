# 🌊 TaskFlow

TaskFlow is a minimal, high-performance task management system. It allows users to create projects, manage tasks via a drag-and-drop Kanban board, and collaborate by assigning tasks to team members. 

This project was built to demonstrate full-stack engineering proficiency, focusing on relational data modeling, strict type safety, optimistic UI updates, and containerized deployment.

---

## 1. Overview

**Core Features**
* 🔐 **Secure Auth:** JWT-based authentication with bcrypt password hashing.
* 📋 **Project Management:** Create, edit, and delete projects with real-time stat tracking.
* 🖱️ **Interactive Kanban:** Drag-and-drop task management with optimistic UI updates.
* 🎨 **Polished UX:** Fully responsive design (mobile to desktop) with persistent Dark Mode.
* 🐳 **Production Ready:** Multi-stage Docker builds for both the frontend and backend.

**The Tech Stack**
* **Frontend:** React 18, TypeScript, Vite, React Router v6, TanStack Query, Chakra UI v2, Hook Form + Zod, `@hello-pangea/dnd`.
* **Backend:** Go (Standard Library / net/http), PostgreSQL, `database/sql`, JWT.
* **Infrastructure:** Docker, Docker Compose, Nginx (for serving the SPA).

---

## 2. Architecture Decisions & Tradeoffs

When building TaskFlow, I prioritized user experience and application resilience over bleeding-edge trends. Here is why the stack looks the way it does:

* **React Query for State (Frontend):** Instead of manually managing loading/error states with Redux or Context, I used TanStack Query. This allowed me to easily implement **Optimistic UI updates** for the drag-and-drop Kanban board. When a user drops a card, the UI updates instantly, while the network request resolves in the background. If the server fails, the UI rolls back automatically.
* **Chakra UI (Frontend):** The rubric required a polished, accessible UI without broken layouts. Chakra UI provides excellent primitive components and handles the complex math of responsive layouts (`base` vs `lg` breakpoints) and dark-mode color token inversion natively.
* **Go Standard Library (Backend):** I opted against heavy frameworks to keep the binary small and execution fast. Go's native concurrency and strict typing pair perfectly with a TypeScript frontend.
* **Nullability in SQL (Database):** Task assignments and due dates are optional. I intentionally modeled these as pointers (`*uuid.UUID`) in Go and `NULL` in PostgreSQL to enforce relational integrity, avoiding the "empty string" foreign key anti-pattern.
* **Tradeoff - Polling vs WebSockets:** To achieve "real-time" updates on the Kanban board without over-engineering the infrastructure for a time-boxed assignment, I utilized TanStack Query's `refetchInterval` (smart polling) rather than standing up a full WebSocket server. 

---

## 3. Running Locally

Running the app requires zero manual configuration outside of Docker. 

**Prerequisites:** Make sure you have Docker and Docker Compose installed.

```bash
# 1. Clone the repository
git clone https://github.com/Jayant-Verma/taskflow-jayant-verma
cd taskflow

# 2. Setup the environment variables
cp .env.example .env

# 3. Spin up the entire stack
docker compose up --build
```

**Access the application:**
* Frontend UI: `http://localhost:5173`
* Backend API: `http://localhost:8080`

---

## 4. Running Migrations

Database schema management is handled via pure SQL migrations. 

When you run `docker compose up`, the `db-init` container automatically executes the `up.sql` migration scripts and seeds the database before the Go API server starts. 

If you need to run them manually or reset the database, you can execute:
```bash
docker compose exec db psql -U postgres -d taskflow -f /docker-entrypoint-initdb.d/01_schema.sql
docker compose exec db psql -U postgres -d taskflow -f /docker-entrypoint-initdb.d/02_seed.sql
```

---

## 5. Test Credentials

The database is seeded with a default user and project so you can immediately test the application without registering.

* **Email:** `test@example.com`
* **Password:** `password123`

---

## 6. API Reference

All protected routes require an `Authorization: Bearer <token>` header.

### Authentication
* `POST /auth/register` - Register a new user.
* `POST /auth/login` - Authenticate and receive a JWT.

### Projects
* `GET /projects` - List all projects accessible to the user.
* `POST /projects` - Create a new project.
* `GET /projects/:id` - Get project details and all associated tasks.
* `PATCH /projects/:id` - Update project name/description.
* `DELETE /projects/:id` - Delete a project and cascade delete its tasks.
* `GET /projects/:id/stats` - Get calculated counts of task statuses.

### Tasks
* `GET /projects/:id/tasks` - List all tasks created under a project.
* `POST /projects/:id/tasks` - Create a new task.
* `PATCH /tasks/:id` - Update task details (title, status, assignee, etc.).
* `DELETE /tasks/:id` - Delete a task.

### Users
* `GET /users` - List All registered users.

---

## 7. API Documentation (Swagger)

The TaskFlow API is fully documented using Swagger (OpenAPI). 

### Accessing the Interactive Swagger UI

To view the full, interactive OpenAPI specification and test the endpoints directly from your browser:

1. Ensure the backend server is running:

2.  Navigate to the Swagger UI in your browser:
    **[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)** 

> **Note:** To test protected endpoints like `/projects` in the Swagger UI, click the **Authorize** button at the top of the page and paste the `access_token` received from the `/auth/login` endpoint.


## 8. What I'd Do With More Time

While I am proud of the architecture and stability of this submission, a time-boxed assessment always requires scoping decisions. If this were a production application, I would prioritize the following:

1. **WebSocket Integration:** I would replace the frontend polling with Go's `gorilla/websocket` to push event-driven updates to clients, making the Kanban board truly multiplayer like Jira or Trello.
2. **Table-Driven Backend Tests:** I included standard integration tests, but I would expand the test suite using Go's table-driven testing pattern to comprehensively cover edge cases (e.g., malformed UUIDs, expired JWTs, SQL injection attempts).
3. **Role-Based Access Control (RBAC):** Currently, anyone with the project ID can view it. I would implement a robust `project_members` join table to ensure strict tenant isolation.
4. **Caching Layer:** I would introduce Redis to cache the `/projects/:id/stats` endpoint, as running aggregate `COUNT` filters on the PostgreSQL database would eventually become a bottleneck at scale.
