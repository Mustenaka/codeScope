package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BridgeRegistryReader interface {
	ConnectedSessionIDs() []string
}

type BridgeHandler struct {
	registry BridgeRegistryReader
}

func NewBridgeHandler(registry BridgeRegistryReader) *BridgeHandler {
	return &BridgeHandler{registry: registry}
}

type bridgeStatusResponse struct {
	ConnectedSessions []string `json:"connected_sessions"`
	Count             int      `json:"count"`
}

func (h *BridgeHandler) Status(c *gin.Context) {
	connected := h.registry.ConnectedSessionIDs()
	c.JSON(http.StatusOK, bridgeStatusResponse{
		ConnectedSessions: connected,
		Count:             len(connected),
	})
}
