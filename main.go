package main

import (
	"net/http"

	loadenv "github.com/caarlos0/env/v11"
	cshAuth "github.com/computersciencehouse/csh-auth"
	. "github.com/computersciencehouse/selections/internal"
	"github.com/gin-gonic/gin"
)

func main() {
	//Get environment variables into struct
	env := GetEnv()
	loadenv.Parse(env)
	auth := cshAuth.CSHAuth{}
	auth.Init(env.OIDCClientID,
		env.OIDCClientSecret,
		env.SessionSecret,
		env.SessionState,
		env.BaseUri,
		env.BaseUri+"/auth/callback",
		env.BaseUri+"/auth/login",
		[]string{"profile", "email", "groups"},
	)

	//DB Initialization
	InitData()

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	//Define routes here
	router.GET("/auth/login", auth.AuthRequest)
	router.GET("/auth/callback", auth.AuthCallback)
	router.GET("/auth/logout", auth.AuthLogout)
	router.GET("/", auth.AuthWrapper(func(c *gin.Context) {
		cl, _ := c.Get("cshauth")
		user := cl.(cshAuth.CSHClaims).UserInfo
		c.HTML(http.StatusOK, "homepage.tmpl", gin.H{"Username": user.Username})
	}))
	router.GET("/application/upload", auth.AuthWrapper(HandleApplicationUploadPage))
	router.POST("/application/upload", auth.AuthWrapper(HandleApplicationFileUpload))
	router.GET("/session/manage", auth.AuthWrapper(HandleSessionManagementPage))

	router.Run()
}
