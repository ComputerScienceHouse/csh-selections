package internal

import (
	"github.com/google/uuid"
)

type Team struct {
	ID      uuid.UUID `gorm:"primarykey"`
	members []string
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

func getTeamForMember(member string) Team {
	membership := Membership{Member: member}
	res := db.Limit(1).First(&membership)
	if res.Error != nil {
		return Team{}
	}
	team := getTeamByID(membership.TeamID)
	return team
}

func (t Team) getTeamMembership() {
	var memberships []Membership
	db.Find(&memberships, t.ID)
	for _, m := range memberships {
		t.members = append(t.members, m.Member)
	}
}

func (t Team) addMemberToTeam(member string) {
	membership := Membership{TeamID: t.ID, Member: member}
	db.Save(&membership)
}

func getTeamByID(ID uuid.UUID) Team {
	t := Team{}
	db.First(&t, ID)
	t.getTeamMembership()
	return t
}

/* ==============
WEB FUNCTIONS GO HERE
			============= */
