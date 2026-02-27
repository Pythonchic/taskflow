// internal/handlers/login_page.go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// GET /login - показать страницу
func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"status": "ok"})
}
