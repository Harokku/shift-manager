package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"shift-manager/api"
	"shift-manager/db"
)

func main() {
	// Heroku port from env variable
	port := os.Getenv("PORT")
	fmt.Printf("port set to %v\n", port)

	// -----------------------
	// Database connection config
	// -----------------------

	// Heroku Postgres connection and ping
	dbConn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	checkErrorAndPanic(err)

	defer dbConn.Close()

	err = dbConn.Ping()
	checkErrorAndPanic(err)
	fmt.Println("Correctly pinged DB")
	// Create a new db service to interact with Heroku's DB
	dbService := db.Service{Db: dbConn}

	// -----------------------
	// Echo server definition
	// -----------------------

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339}, method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// -----------------------
	// Static definition
	// -----------------------
	e.File("/favicon.ico", "static/favicon.ico")

	e.GET("/ping", func(context echo.Context) error {
		return context.JSON(http.StatusOK, "Pong")
	})

	// -----------------------
	// Routes
	// -----------------------

	// Login route
	e.POST("/login", api.Login(&dbService))

	// Admin group (req auth)
	admin := e.Group("/admin", middleware.JWT([]byte(os.Getenv("SECRET"))))
	admin.GET("", func(context echo.Context) error {
		return context.String(http.StatusNoContent, "Admin route root")
	})
	admin.POST("/passwordreset", api.ResetPwd(&dbService))

	// Users group (req auth)
	users := e.Group("/users", middleware.JWT([]byte(os.Getenv("SECRET"))))
	users.GET("", func(context echo.Context) error {
		return context.String(http.StatusNoContent, "Users management route root")
	})
	users.GET("/userdetails", api.GetUserDetailsFromClaims(&dbService))

	// Gsheet group (req auth)
	gSheet := e.Group("/sheets", middleware.JWT([]byte(os.Getenv("SECRET"))))
	gSheet.GET("", func(context echo.Context) error {
		return context.String(http.StatusNoContent, "Google Sheets route root")
	})
	gSheet.POST("/shift", api.PostShift())

	// -----------------------
	// Server Start
	// -----------------------

	e.Logger.Fatal(e.Start(":" + port))
}

// Default error check with fatal if err != nil
func checkErrorAndPanic(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
