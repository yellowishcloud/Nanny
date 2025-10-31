package routes

import (
	"database/sql"
	"nanny-backend/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// чтобы фронт мог обращаться к твоему API
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// вот эти три эндпоинта ↓↓↓
	r.POST("/register/owner", handlers.RegisterOwner(db))
	r.POST("/register/sitter", handlers.RegisterSitter(db))
	r.POST("/login", handlers.Login(db))

	return r
}
