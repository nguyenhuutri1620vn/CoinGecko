package main

import (
	"interview_test/src/app/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.SetTrustedProxies([]string{"https://api.coingecko.com"})
	historyController := controllers.NewHistoryController()
	router.POST("/get_histories", historyController.GetHistories)

	router.Run(":8080")
}
