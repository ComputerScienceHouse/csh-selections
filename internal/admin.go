package internal

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type RLAuthCode struct {
	ID   int `gorm:"primary_key"`
	Code string
}

func getResLifeCode() string {
	code, b := goCache.Get("RLAuthCode")
	if b {
		return code.(string)
	}
	res := &RLAuthCode{}
	tx := db.Find(res)
	if tx.RowsAffected == 0 {
		res.Code = newResLifeCode()
	}
	goCache.SetDefault("RLAuthCode", res.Code)
	return res.Code
}

func newResLifeCode() string {
	code := rand.Text()
	save := RLAuthCode{Code: code}
	db.Save(&save)
	return code
}

func delResLifeCode() {
	db.Where("1 = 1").Delete(&RLAuthCode{})
	goCache.Delete("RLAuthCode")
}

func HandleAdminPage(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}
	c.HTML(http.StatusOK, "admin.tmpl", templateHeaders(c, map[string]interface{}{"ResLifeCode": getResLifeCode()}))
}

func HandleAdminNewRLCode(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}
	delResLifeCode()
	c.JSON(http.StatusOK, gin.H{})
}

func HandleDebugPage(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}
	obj := listBucketApplications()
	for _, v := range obj {
		fmt.Println(*v.Key, v.Size, *v.LastModified)
	}
	c.HTML(http.StatusOK, "adminDebug.tmpl", templateHeaders(c))
}

func HandleDebugPost(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}
	switch c.PostForm("debugAction") {
	case "ClearBucket":
		clearBucket()
		break
	case "ClearCache":
		goCache = cache.New(5*time.Minute, 10*time.Minute)
	}
	c.JSON(http.StatusOK, nil)
}

func HandleDebugError(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "You wanted an error"})
}

func HandleGetRlLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "rlLogin.tmpl", gin.H{})
}

func HandlePostRlLogin(c *gin.Context) {
	// set cookie
	code := c.PostForm("authcode")
	if code != getResLifeCode() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	c.SetCookie("RLAuth", code, 3600, "", "", false, false)
	c.Redirect(http.StatusFound, "/application/manage")
}
