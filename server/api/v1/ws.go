package v1

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type wsServer interface {
	AddClient(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

// HandleWs - handles websocket requests from the peer.
func HandleWs(ctx context.Context, server wsServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := server.AddClient(ctx, c.Writer, c.Request); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)

			return
		}
	}
}
