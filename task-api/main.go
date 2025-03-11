package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var db *sql.DB

// Task struct represents a task in the task management system
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func main() {
	var err error
	// Establish a connection to the PostgreSQL database
	connStr := "user=postgres password=yourpassword dbname=task_management sslmode=disable" // Update with your password
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Unable to connect to the database:", err)
	}

	// Ensure the connection is valid
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping the database:", err)
	}

	// Setup routes
	http.HandleFunc("/tasks", tasksHandler) // Handles both GET and POST requests for /tasks
	http.HandleFunc("/tasks/", taskHandler) // Handles individual tasks based on ID

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Handler to fetch all tasks and create new tasks
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Fetch all tasks
		tasks := []Task{}
		rows, err := db.Query("SELECT id, title, description, status, created_at, updated_at FROM tasks")
		if err != nil {
			http.Error(w, "Could not fetch tasks", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var task Task
			err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
			if err != nil {
				http.Error(w, "Could not scan task", http.StatusInternalServerError)
				return
			}
			tasks = append(tasks, task)
		}

		// Respond with the task data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)

	case http.MethodPost:
		// Create a new task
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Insert the new task into the database
		err = db.QueryRow(
			"INSERT INTO tasks (title, description, status) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at",
			task.Title, task.Description, task.Status,
		).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			http.Error(w, "Could not insert task", http.StatusInternalServerError)
			return
		}

		// Respond with the created task
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)

	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// Handler for individual tasks (GET, PUT, DELETE)
func taskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/tasks/"):]

	switch r.Method {
	case http.MethodGet:
		// Fetch a specific task
		var task Task
		err := db.QueryRow("SELECT id, title, description, status, created_at, updated_at FROM tasks WHERE id = $1", id).Scan(
			&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		// Respond with the task data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)

	case http.MethodPut:
		// Update task status
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE tasks SET status = $1 WHERE id = $2", task.Status, id)
		if err != nil {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content

	case http.MethodDelete:
		// Delete a task
		_, err := db.Exec("DELETE FROM tasks WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content

	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
