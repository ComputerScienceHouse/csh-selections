package internal

import (
	cshAuth "github.com/computersciencehouse/csh-auth"
	"github.com/gin-gonic/gin"
)

func templateHeaders(c *gin.Context, data ...map[string]any) gin.H {
	cl, _ := c.Get("cshauth")
	user := cl.(cshAuth.CSHClaims).UserInfo
	ret := gin.H{"Username": user.Username}
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
