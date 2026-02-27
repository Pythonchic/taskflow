package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func TasksPage (c *gin.Context) {
    c.HTML(http.StatusOK, "tasks.html", gin.H{"status": "ok"})
}
