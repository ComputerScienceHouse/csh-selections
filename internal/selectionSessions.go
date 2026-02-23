package internal

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionState struct {
	PK           int `gorm:"primarykey"`
	State        bool
	User         string
	ModifiedDate time.Time
}

func IsSelectionsActive() bool {
	state := SessionState{}
	db.Last(&state)
	state, err := gorm.G[SessionState](db).Last(context.Background())
	if err != nil {
		log.Println("Error checking Selections active state:", err)
	}
	return state.State
}

func setSelectionsState(state bool, member string) {
	currentState := IsSelectionsActive()
	if currentState == state {
		log.Println("Selections State was requested to be", state, "but is already", currentState)
	}
	log.Println("Selections State was requested to be", state, "by", member)
	newState := SessionState{State: state, User: member, ModifiedDate: time.Now()}
	db.Save(&newState)
}

func HandleSessionHomePage(c *gin.Context) {
	user := getUserData(c)
	addedToTeam := getTeamForMember(user.Username).ID != uuid.UUID{}
	c.HTML(http.StatusOK, "homepage.tmpl", templateHeaders(c, map[string]any{"OnATeam": addedToTeam}))
}

func HandleSessionMemberList(c *gin.Context) {
	var members []string
	for _, user := range oidcClient.GetActiveUsers() {
		members = append(members, user.Username)
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

func HandleSessionManagementPage(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
	c.HTML(http.StatusOK, "sessionManagement.tmpl", templateHeaders(c))
}

func HandleSessionChanging(c *gin.Context) {
	user := getUserData(c)
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
	if res, ok := c.GetPostForm("state"); ok {
		switch res {
		case "start":
			//TODO: don't allow starting with 0 teams
			setSelectionsState(true, user.Username)
			break
		case "stop":
			//TODO: wipe all teams
			setSelectionsState(false, user.Username)
			break
		default:
			c.JSON(http.StatusBadRequest, "Invalid state to change to")
		}
	}
	HandleSessionManagementPage(c)
}
