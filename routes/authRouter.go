package routes

import (
	controller "github.com/Kaustubh8691/golang-backend/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("user/signup", controller.Signup())
	incomingRoutes.POST("user/login", controller.Login())
	incomingRoutes.GET("/users", controller.GetUser())

}
