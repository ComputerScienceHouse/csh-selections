package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleDebugPage(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}
	listBucketApplications()
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
	}
	c.JSON(http.StatusOK, nil)
}
