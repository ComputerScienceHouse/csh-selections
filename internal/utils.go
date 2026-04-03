package internal

import (
	"net/http"
	"reflect"
	"slices"

	cshAuth "github.com/computersciencehouse/csh-auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func templateHeaders(c *gin.Context, data ...map[string]any) gin.H {
	user := getUserData(c)
	ret := gin.H{"Username": user.Username, "SessionState": IsSelectionsActive(), "IsAdmin": isUserAdmin(c), "NilUUID": uuid.UUID{}}
	for _, d := range data {
		for k, v := range d {
			ret[k] = v
		}
	}
	return ret
}

func Length(item any) int {
	defer func() {
		recover()
	}()
	ret := 0
	ret = reflect.ValueOf(item).Len()
	return ret
}

func getUserData(c *gin.Context) cshAuth.CSHUserInfo {
	cl, b := c.Get("cshauth")
	if !b {
		return cshAuth.CSHUserInfo{
			Username: "ResLifeUser",
			FullName: "ResLife Staff",
			Groups:   []string{"reslife"},
		}
	}
	user := cl.(cshAuth.CSHClaims).UserInfo
	return user
}

func isUserAdmin(c *gin.Context) bool {
	user := getUserData(c)
	return slices.Contains(user.Groups, "eboard-evaluations") ||
		slices.Contains(user.Groups, "active_rtp") ||
		slices.Contains(user.Groups, "reslife")
}

// auth for advisors or admins
func PaulthWrapper(auth cshAuth.CSHAuth, h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// look for CSH auth
		cookie, err := c.Cookie(cshAuth.CookieName)
		if cookie != "" && err == nil {
			auth.AuthWrapper(h)(c)
			return
		}
		// roll our own
		cookie, err = c.Cookie("RLAuth")
		if cookie == "" || err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/rl/login")
			return
		}
		if cookie == getResLifeCode() {
			h(c)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}
