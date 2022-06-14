package main

import (
	"os"

	routes "github.com/Kaustubh8691/golang-backend/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// func CORSMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
// 		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
// 		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
// 		// c.Writer.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
// 		// c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
// 		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
// 		c.Writer.Header().Add("Access-Control-Allow-Credentials", "true")

// 		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
// 		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
// 		c.Writer.Header().Add("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

// 		c.Writer.Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
// 		if c.Request.Method == "OPTIONS" {
// 			c.AbortWithStatus(204)
// 			return
// 		}

// 		c.Next()
// 	}
// }

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(cors.Default())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "access granted"})
	})
	router.GET("/api2", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "access granted in 2"})
	})

	router.Run(":" + port)
}
