package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
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

type SessionAttendance struct {
	Member   string `gorm:"primarykey"`
	IsEboard bool
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

func isMemberAttending(member string) bool {
	tx := db.First(&SessionAttendance{Member: member})
	return tx.RowsAffected > 0
}

func getAttendingMembers() ([]OIDCUser, int) {
	var attendingMembers []SessionAttendance
	db.Order("is_eboard").Find(&attendingMembers)
	eboard := 0
	ret := make([]OIDCUser, len(attendingMembers))
	for i, member := range attendingMembers {
		ret[i] = *oidcClient.GetUserInfo(member.Member)
		if ret[i].IsEboard() {
			eboard++
		}
	}
	return ret, eboard
}

// Page functions

func HandleSessionHomePage(c *gin.Context) {
	user := getUserData(c)
	addedToTeam := getTeamForMember(user.Username).ID != uuid.UUID{}
	c.HTML(http.StatusOK, "homepage.tmpl", templateHeaders(c, map[string]any{"OnATeam": addedToTeam, "IsAttending": isMemberAttending(user.Username)}))
}

func HandleSessionManagementPage(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
	attendees, eboard := getAttendingMembers()
	c.HTML(http.StatusOK, "sessionManagement.tmpl", templateHeaders(c, map[string]any{"Attendance": attendees, "EBoard": eboard, "Teams": getAllTeams()}))
}

// POST functions

func HandleSessionAddMembers(c *gin.Context) {
	members := c.PostForm("membersSelect")
	memberList := strings.Split(members, ",")
	for _, member := range memberList {
		fmt.Println(member)
		user := oidcClient.GetUserInfo(member)
		save := SessionAttendance{Member: member, IsEboard: user.IsEboard()}
		tx := db.Create(&save)
		if tx.Error != nil {
			log.Println("HandleSessionAddMembers", tx.Error)
		}
	}
	HandleSessionManagementPage(c)
}

func HandleSessionEligibleMemberList(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
	var members []map[string]string
	for _, user := range oidcClient.GetActiveUsers() {
		if isMemberAttending(user.Username) {
			continue
		}
		members = append(members, map[string]string{"Username": user.Username, "Name": user.FirstName + " " + user.LastName})
	}
	c.JSON(http.StatusOK, gin.H{"Members": members})
}

func HandleSessionAttendingList(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
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
