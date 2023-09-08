package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
)

var db *gorm.DB

type User struct {
	ID       uint   `gorm:"primary_key"`
	Email    string `gorm:"unique_index"`
	Password string
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

func main() {
	// Initialize the database
	initDB("zadanko.db")

	// Create a new Gin router
	router := gin.Default()

	// Define API routes
	router.POST("/login", loginHandler)
	router.GET("/users", listUsersHandler)
	router.PUT("/change-password/:id", changePasswordHandler)

	// Start the HTTP server on a specified port (e.g., 8080)
	port := "8080"
	router.Run(":" + port)

	// Check for CLI commands
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "add-user":
			if len(os.Args) != 4 {
				fmt.Println("Usage: ./zadanko <email> <password>")
			} else {
				email := os.Args[2]
				password := os.Args[3]
				addUser(email, password)
			}
		default:
			fmt.Println("Unknown command:", command)
		}
	}
}

func initDB(dsn string) (*gorm.DB, error) {
	var err error
	db, err = gorm.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&User{})
	return db, nil
}

func loginHandler(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var storedUser User
	if err := db.Where("email = ?", request.Email).First(&storedUser).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(request.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully"})
}

func listUsersHandler(c *gin.Context) {
	var users []User
	if err := db.Select("id, email").Find(&users).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, UserResponse{ID: user.ID, Email: user.Email})
	}

	c.JSON(http.StatusOK, gin.H{"users": userResponses})
}

func changePasswordHandler(c *gin.Context) {
	var request ChangePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	var storedUser User
	if err := db.Where("id = ?", id).First(&storedUser).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(request.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password does not match"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while hashing the password"})
		return
	}

	storedUser.Password = string(hashedPassword)
	db.Save(&storedUser)

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func addUser(email, password string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error while hashing the password")
		return
	}

	user := User{
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := db.Create(&user).Error; err != nil {
		fmt.Println("Error while adding the user to the database")
		return
	}

	fmt.Println("User added successfully")
}
