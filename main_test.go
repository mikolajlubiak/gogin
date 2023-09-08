package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = initDB(":memory:") // use in-memory database for testing
	if err != nil {
		panic("failed to connect to database")
	}

	code := m.Run()

	testDB.Close() // close connection after tests are done

	os.Exit(code)
}

func TestAddUser(t *testing.T) {
	addUser("test@example.com", "password")

	var user User
	if err := testDB.Where("email = ?", "test@example.com").First(&user).Error; err != nil {
		t.Fatalf("User not found in the database")
	}

	assert.Equal(t, user.Email, "test@example.com")
}

func TestLoginHandler(t *testing.T) {
	addUser("login@example.com", "password")

	router := gin.Default()
	router.POST("/login", loginHandler)

	body := bytes.NewBufferString(`{"email":"login@example.com", "password":"password"}`)
	req, _ := http.NewRequest("POST", "/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, 200, resp.Code)
	assert.Contains(t, resp.Body.String(), "Logged in successfully")
}

func TestListUsersHandler(t *testing.T) {
	addUser("list@example.com", "password")

	router := gin.Default()
	router.GET("/users", listUsersHandler)

	req, _ := http.NewRequest("GET", "/users", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, 200, resp.Code)
	assert.Contains(t, resp.Body.String(), "list@example.com")
}
