package httpserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"go-playground/internal/ws"
)

// Server holds the Gin engine and configuration.
type Server struct {
	engine *gin.Engine
}

// New constructs the HTTP server with routes and middleware.
func New() *Server {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	hub := ws.NewHub()
	go hub.Run()

	engine.POST("/notify/user", func(c *gin.Context) {
		// Accept a user id and message, then broadcast to all connections for that user.
		userID := c.Query("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
			return
		}
		message := c.Query("message")
		if message == "" {
			message = "notification"
		}
		hub.SendToUser(userID, []byte(message))
		c.JSON(http.StatusOK, gin.H{"status": "sent"})
	})

	engine.POST("/notify/redis", func(c *gin.Context) {
		// Publish to per-user Redis channels so other nodes can fan out.
		rawIDs := c.Query("ids")
		if rawIDs == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing ids"})
			return
		}
		message := c.Query("message")
		if message == "" {
			message = "redis notification"
		}
		parts := strings.Split(rawIDs, ",")
		userIDs := make([]string, 0, len(parts))
		for _, part := range parts {
			id := strings.TrimSpace(part)
			if id != "" {
				userIDs = append(userIDs, id)
			}
		}
		if len(userIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no valid ids"})
			return
		}
		hub.PublishToUsers(userIDs, []byte(message))
		c.JSON(http.StatusOK, gin.H{"status": "sent", "count": len(userIDs)})
	})

	engine.GET("/ws", func(c *gin.Context) {
		ws.HandleWebSocket(c.Writer, c.Request, hub)
	})

	return &Server{engine: engine}
}

// Run starts the Gin server on the provided port.
func (s *Server) Run(port int) error {
	return s.engine.Run(fmt.Sprintf(":%d", port))
}
