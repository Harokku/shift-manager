package main

import (
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"shift-manager/api"
	"shift-manager/db"
)

// -----------------------
// Custom middleware
// -----------------------

// checkIfRole check if passed JWT contain (r) role and forward or stop the request
//
// Don't check if JWT is valid or present, but assume it's true,
// chain with echo's middleware.JWT to avoid problems
func checkIfRole(r string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Read user from JWT and extract claims
			user := c.Get("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)
			roles := claims["role"].([]interface{})

			// Check if roles contain (r)
			for _, role := range roles {
				// if true, call next middleware
				if role == r {
					return next(c)
				}
			}
			// if role is not found, drop request and return not authorized error
			return echo.ErrUnauthorized
		}
	}
}

// -----------------------
// Main
// -----------------------

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

	// Admin group (req auth and admin role)
	admin := e.Group("/admin", middleware.JWT([]byte(os.Getenv("SECRET"))))
	admin.Use(checkIfRole("admin"))
	admin.GET("", func(context echo.Context) error {
		return context.String(http.StatusNoContent, "Admin route root")
	})
	admin.POST("/passwordreset", api.ResetPwd(&dbService))

	// Manager group (req auth and manager role)
	manager := e.Group("/manager", middleware.JWT([]byte(os.Getenv("SECRET"))))
	manager.Use(checkIfRole("manager"))
	manager.PUT("/dochange", api.PutChange())
	manager.POST("/managechange", api.ManageChangeRequest(&dbService))

	// Users group (req auth)
	users := e.Group("/users", middleware.JWT([]byte(os.Getenv("SECRET"))))
	users.GET("", func(context echo.Context) error {
		return context.String(http.StatusNoContent, "Users management route root")
	})
	users.GET("/all", api.GetAllUserNames(&dbService))
	users.GET("/userdetails", api.GetUserDetailsFromClaims(&dbService))

	// Shift data (req auth)
	shiftData := e.Group("/shiftdata", middleware.JWT([]byte(os.Getenv("SECRET"))))
	shiftData.GET("/all", api.GetAllFormData(&dbService))
	shiftData.GET("/today", api.GetLoggedInOperatorShift())
	shiftData.GET("/date/:date", api.GetLoggedInOperatorShiftByDate())

	// Change request (req auth)
	changeRequest := e.Group("/changes", middleware.JWT([]byte(os.Getenv("SECRET"))))
	changeRequest.POST("/request", api.RequestChange(&dbService))
	changeRequest.GET("/all", api.GetAllChanges(&dbService), checkIfRole("manager"))
	changeRequest.GET("/user", api.GetAllChangesForUser(&dbService))

	// License request (req auth)
	licenseRequest := e.Group("/license", middleware.JWT([]byte(os.Getenv("SECRET"))))
	licenseRequest.POST("/request", api.PostLicense())

	// Gsheet group (req auth)
	gSheet := e.Group("/sheets", middleware.JWT([]byte(os.Getenv("SECRET"))))
	gSheet.GET("", func(context echo.Context) error {
		return context.String(http.StatusNoContent, "Google Sheets route root")
	})
	gSheet.POST("/shift", api.PostShift())
	gSheet.GET("/pastshifts", api.GetPostedShifts())

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
