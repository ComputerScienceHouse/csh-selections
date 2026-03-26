package internal

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Team struct {
	ID            uuid.UUID  `gorm:"primarykey"`
	ApplicationID uuid.UUID  `gorm:"refrences:Application,ID"`
	Members       []OIDCUser `gorm:"-"`
}

type Membership struct {
	TeamID uuid.UUID `gorm:"primarykey;references:TeamID,ID"`
	Member string    `gorm:"primarykey"`
}

/* ==============
ACTUAL FUNCTIONS GO HERE
			============= */

func createTeam() Team {
	ret := Team{ID: uuid.New()}
	db.Create(&ret)
	return ret
}

func getTeamForMember(member string) *Team {
	membership := Membership{Member: member}
	res := db.Limit(1).First(&membership)
	if res.Error != nil {
		return &Team{}
	}
	team := getTeamByID(membership.TeamID)
	return team
}

func isMemberOnTeam(member string) bool {
	return getTeamForMember(member).ID != uuid.UUID{}
}

func (t *Team) getTeamMembership() {
	var memberships []Membership
	db.Find(&memberships, t.ID)
	for _, m := range memberships {
		t.Members = append(t.Members, *oidcClient.GetUserInfo(m.Member))
	}
}

func (t *Team) deleteTeam() {
	db.Delete(t)
}

func (t *Team) addMemberToTeam(member OIDCUser) {
	membership := Membership{TeamID: t.ID, Member: member.Username}
	t.Members = append(t.Members, member)
	db.Save(&membership)
}

func (t *Team) addMembersToTeam(members []OIDCUser) {
	for _, member := range members {
		t.addMemberToTeam(member)
	}
}

func removeMemberFromTeam(member OIDCUser) {
	res := Membership{}
	db.Where("member = ?", member.Username).First(&res)
	db.Delete(&res)
	tx := db.Where(&res.TeamID).Find(nil)
	if tx.RowsAffected == 0 {
		getTeamByID(res.TeamID).deleteTeam()
	}
}

func (t *Team) setTeamApplication(appID uuid.UUID) {
	t.ApplicationID = appID
	db.Save(&t)
}

func getAllTeams() []*Team {
	var teams []*Team
	db.Find(&teams)
	for _, team := range teams {
		team.getTeamMembership()
	}
	return teams
}

func getTeamByID(ID uuid.UUID) *Team {
	t := &Team{}
	db.First(&t, ID)
	t.getTeamMembership()
	return t
}

func dropAllTeams() {
	tx := db.Where("1 = 1").Delete(&Membership{})
	fmt.Println(tx.RowsAffected, tx.Error)
	db.Where("1 = 1").Delete(&Team{})
}

func disperseMembersToTeams(members []OIDCUser) {
	teams := getAllTeams()
	teamCount := len(teams)
	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})
	for index, member := range members {
		if isMemberOnTeam(member.Username) {
			continue
		}
		currentTeam := index % teamCount
		teams[currentTeam].addMemberToTeam(member)
	}
}

/* ==============
WEB FUNCTIONS GO HERE
			============= */

func HandleTeamCreation(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
	dropAllTeams()
	teamCountRaw := c.PostForm("teamCount")
	teamCount, err := strconv.Atoi(teamCountRaw)
	if err != nil {
		log.Println("Error converting teamCount to int:", err)
		return
	}
	teams := make([]Team, teamCount)
	members, eboardCount := getAttendingMembers()
	if teamCount > eboardCount {
		log.Println("teamCount larger than eboardCount, falling back to eboardCount", eboardCount)
		teamCount = eboardCount
	}
	eboard := members[0:teamCount]
	members = members[teamCount:]
	for i := 0; i < teamCount; i++ {
		if isMemberOnTeam(eboard[i].Username) {
			continue
		}
		teams[i] = createTeam()
		teams[i].addMemberToTeam(eboard[i])
	}
	disperseMembersToTeams(members)
	c.Status(200)
}

func HandleTeamApplicationAssignment(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, "You're not authorized to access this page!")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Println("Failed while parsing team id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	appId, err := uuid.Parse(c.PostForm("applicationID"))
	if err != nil {
		log.Println("Failed while parsing application id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	getTeamByID(id).setTeamApplication(appId)
	app := getApplication(appId)
	app.Assigned = true
	db.Save(&app)
}
