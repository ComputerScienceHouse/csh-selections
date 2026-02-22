package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SessionState struct {
	pk           int `gorm:"primarykey"`
	State        bool
	User         string
	ModifiedDate time.Time
}

func isSelectionsActive() bool {
	state, err := gorm.G[SessionState](db).Last(context.Background())
	if err != nil {
		log.Println("Error checking Selections active state:", err)
	}
	return state.State
}

func HandleSessionManagementPage(c *gin.Context) {
	user := getUserData(c)
	fmt.Println(user.Groups)
	if !slices.Contains(user.Groups, "eboard-evaluations") && !slices.Contains(user.Groups, "active_rtp") {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
	c.HTML(http.StatusOK, "session_management.tmpl", templateHeaders(c))
}
