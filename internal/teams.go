package internal

import (
	"github.com/google/uuid"
)

type Team struct {
	ID      uuid.UUID `gorm:"primarykey"`
	Members []string
}

/* ==============
ACTUAL FUNCTIONS GO HERE
			============= */

func createTeam() Team {
	ret := Team{ID: uuid.New()}
	db.Create(&ret)
	return ret
}

func (t Team) addMemberToTeam(member string) {
	t.Members = append(t.Members, member)
	db.Save(&t)
}

func getTeamByID(ID uuid.UUID) Team {
	t := Team{}
	db.First(&t, ID)
	return t
}

/* ==============
WEB FUNCTIONS GO HERE
			============= */
