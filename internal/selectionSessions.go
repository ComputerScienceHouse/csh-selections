package internal

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

func removeMemberFromAttending(member string) {
	user := oidcClient.GetUserInfo(member)
	del := SessionAttendance{Member: user.Username}
	if isMemberOnTeam(user.Username) {
		removeMemberFromTeam(*user)
	}
	db.Delete(&del)
}

func dropAllAttendance() {
	db.Where("1 = 1").Delete(&SessionAttendance{})
}

func getAttendingMembers() ([]OIDCUser, int) {
	var attendingMembers []SessionAttendance
	db.Order("is_eboard DESC").Find(&attendingMembers)
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

func getUnassignedAttendees() []OIDCUser {
	var unassignedMembers []SessionAttendance
	db.Where("member NOT IN (SELECT member from memberships)").Find(&unassignedMembers)
	ret := make([]OIDCUser, len(unassignedMembers))
	for i, member := range unassignedMembers {
		ret[i] = *oidcClient.GetUserInfo(member.Member)
		if ret[i].IsEboard() {
		}
	}
	return ret
}

func tryAddingUnassignedMember(team *Team) {
	if unass := getUnassignedAttendees(); len(unass) > 0 {
		team.addMemberToTeam(unass[0])
	}
}

func clearSession() bool {
	if IsSelectionsActive() {
		log.Println("Selections State was requested to cleared but selections is active. This will not continue.")
		return false
	}
	dropAllTeams()
	db.Where("1 = 1").Delete(&SessionAttendance{})
	//TODO: delete S3 objects (application PDFs)
	return true
}

// Page functions

func HandleSessionHomePage(c *gin.Context) {
	user := getUserData(c)
	team := getTeamForMember(user.Username)
	c.HTML(http.StatusOK, "homepage.tmpl", templateHeaders(c, map[string]any{
		"OnATeam":         isMemberOnTeam(user.Username),
		"IsAttending":     isMemberAttending(user.Username),
		"Team":            team,
		"TeamApplication": getApplication(team.ApplicationID),
		"Css":             "homepage.css",
		"Criteria":        getCriteria(),
		"Rated":           didUserRateApplication(team.ApplicationID, user.Username),
	}))
}

func HandleSessionManagementPage(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
	attendees, eboard := getAttendingMembers()
	teams := getAllTeams()
	c.HTML(http.StatusOK, "sessionManagement.tmpl", templateHeaders(c, map[string]any{
		"Attendance": attendees,
		"EBoard":     eboard,
		"Teams":      teams,
		"Unassigned": getUnassignedAttendees(),
	}))
}

// POST functions

func HandleSessionAddMembers(c *gin.Context) {
	members := c.PostForm("membersSelect")
	members = strings.TrimSpace(members)
	if members == "" {
		c.JSON(http.StatusBadRequest, "You need to provide members")
		return
	}
	memberList := strings.Split(members, ",")
	for _, member := range memberList {
		user := oidcClient.GetUserInfo(member)
		save := SessionAttendance{Member: member, IsEboard: user.IsEboard()}
		tx := db.Create(&save)
		if tx.Error != nil {
			log.Println("HandleSessionAddMembers", tx.Error)
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{})
}

func HandleSessionRemoveMember(c *gin.Context) {
	json := map[string]string{"member": ""}
	err := c.ShouldBindJSON(&json)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	member := strings.TrimSpace(json["member"])
	if member == "" {
		c.JSON(http.StatusBadRequest, "You need to provide a member")
		return
	}
	removeMemberFromAttending(member)
	c.JSON(http.StatusOK, gin.H{})
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
	c.JSON(http.StatusOK, members)
}

func HandleSessionChanging(c *gin.Context) {
	user := getUserData(c)
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You're not authorized to access this page!"})
		return
	}
	if res, ok := c.GetPostForm("state"); ok {
		switch res {
		case "start":
			if len(getAllTeams()) < 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot start with 0 teams!"})
				return
			}
			setSelectionsState(true, user.Username)
			break
		case "stop":
			dropAllTeams()
			dropAllAttendance()
			dropAllApplications()

			setSelectionsState(false, user.Username)
			break
		default:
			c.JSON(http.StatusBadRequest, "Invalid state to change to")
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, "Invalid state")
		return
	}
	c.JSON(http.StatusOK, nil)
}

func HandleClearingSession(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
	if !clearSession() {
		c.JSON(http.StatusBadRequest, "Selections cannot be cleared while active.")
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
