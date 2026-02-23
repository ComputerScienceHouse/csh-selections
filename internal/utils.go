package internal

import (
	"slices"

	cshAuth "github.com/computersciencehouse/csh-auth"
	"github.com/gin-gonic/gin"
)

func templateHeaders(c *gin.Context, data ...map[string]any) gin.H {
	user := getUserData(c)
	ret := gin.H{"Username": user.Username, "SessionState": IsSelectionsActive(), "IsAdmin": isUserAdmin(c)}
	for _, d := range data {
		for k, v := range d {
			ret[k] = v
		}
	}
	return ret
}

func getUserData(c *gin.Context) cshAuth.CSHUserInfo {
	cl, _ := c.Get("cshauth")
	user := cl.(cshAuth.CSHClaims).UserInfo
	return user
}

func isUserAdmin(c *gin.Context) bool {
	user := getUserData(c)
	return slices.Contains(user.Groups, "eboard-evaluations") || slices.Contains(user.Groups, "active_rtp")
}
