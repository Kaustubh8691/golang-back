package routes

import (
	controller "github.com/Kaustubh8691/golang-backend/controllers"
	"github.com/Kaustubh8691/golang-backend/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	// incomingRoutes.GET("/users", controller.GetUser())
	incomingRoutes.Use(middleware.Authenticate())

	incomingRoutes.GET("/user/:user_id", controller.GetUser())
	incomingRoutes.GET("/data", controller.GetData())
	incomingRoutes.POST("/datas", controller.Crea())
	// incomingRoutes.GET("/data/:user_id", controller.GetData())
	incomingRoutes.PUT("/data/:user_id", controller.UpdateData())
	incomingRoutes.DELETE("/data/:user_id", controller.DeleteData())

}
