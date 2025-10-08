package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Task struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Category string `json:"category"`
}

var db *gorm.DB

func main() {
	dsn := "host=localhost user=postgres password=arthur123 dbname=TaskGo port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	if err := db.AutoMigrate(&Task{}); err != nil {
		log.Fatal("Migration failed:", err)
	}

	r := gin.Default()

	r.GET("/task/:id", getTask)
	r.GET("/tasks", getAllTasks)
	r.POST("/task", createTask)
	r.PUT("/task/:id", updateTask)       // Полное обновление
	r.PATCH("/task/:id", updateCategory) // Только категория
	r.DELETE("/task/:id", deleteTask)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(r.Run(":" + port))
}

func createTask(c *gin.Context) {
	var t Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func getAllTasks(c *gin.Context) {
	var tasks []Task
	if err := db.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func getTask(c *gin.Context) {
	id := c.Param("id")
	var t Task
	if err := db.First(&t, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var existing Task
	if err := db.First(&existing, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	var newTask Task
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.Title = newTask.Title
	existing.Text = newTask.Text
	existing.Category = newTask.Category

	if err := db.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task fully updated"})
}

func updateCategory(c *gin.Context) {
	id := c.Param("id")
	var t Task
	if err := db.First(&t, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	var data struct {
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t.Category = data.Category
	if err := db.Save(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category updated"})
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := db.Delete(&Task{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}
