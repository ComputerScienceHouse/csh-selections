package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

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
