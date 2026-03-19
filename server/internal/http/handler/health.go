package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func Health(appName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{
			Name:   appName,
			Status: "ok",
		})
	}
}
