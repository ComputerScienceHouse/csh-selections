package main

import (
	"fmt"
	"net/http"

	loadenv "github.com/caarlos0/env/v11"
	cshAuth "github.com/computersciencehouse/csh-auth"
	"github.com/computersciencehouse/selections/internal"
	"github.com/gin-gonic/gin"
)

type environment struct {
	OIDCClientID     string `env:"OIDC_CLIENT_ID"`
	OIDCClientSecret string `env:"OIDC_CLIENT_SECRET"`
	SessionSecret    string `env:"SESSION_SECRET"`
	SessionState     string `env:"SESSION_STATE"`
	BaseUri          string `env:"BASE_URI"`
}

var env environment

func main() {
	//Get environment variables into struct
	loadenv.Parse(&env)
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

	fmt.Println(auth)

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
	router.POST("/application/upload", auth.AuthWrapper(internal.HandleApplicationUpload))

	router.Run()
}
