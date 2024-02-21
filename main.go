// main.go

package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Task represents a task entity
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
}

// initializing a variable which has type *sql.db which means pointer to sql.db
// it is a type provided by database/sql
var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		due_date TEXT,
		status TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Error creating tasks table:", err)
	}
}

func main() {
	initDB()
	defer db.Close()

	router := gin.Default()

	router.POST("/tasks", createTask)
	router.GET("/tasks/:id", getTask)
	router.PUT("/tasks/:id", updateTask)
	router.DELETE("/tasks/:id", deleteTask)
	router.GET("/tasks", listTasks)

	err := router.Run(":8080")
	if err != nil {
		log.Fatal("Unable to start server:", err)
	}
}

// c *gin.Context provides methods to access request parameters, headers, body, and other information related to the HTTP request
func createTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("INSERT INTO tasks (title, description, due_date, status) VALUES (?, ?, ?, ?)",
		task.Title, task.Description, task.DueDate, task.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	lastInsertID, err := result.LastInsertId() // Retrieve last insert ID
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	task.ID = int(lastInsertID) // Convert last insert ID to int

	c.JSON(http.StatusCreated, task)
}

func getTask(c *gin.Context) {
	var task Task
	id := c.Param("id")

	err := db.QueryRow("SELECT id, title, description, due_date, status FROM tasks WHERE id = ?", id).
		Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE tasks SET title = ?, description = ?, due_date = ?, status = ? WHERE id = ?",
		task.Title, task.Description, task.DueDate, task.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func listTasks(c *gin.Context) {
	var tasks []Task

	rows, err := db.Query("SELECT id, title, description, due_date, status FROM tasks")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, tasks)
}
