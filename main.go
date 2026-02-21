package main

import (
	loadenv "github.com/caarlos0/env/v11"
	cshAuth "github.com/computersciencehouse/csh-auth"
	"github.com/gin-gonic/gin"
)

type environment struct {
	OIDCClientID     string `env:"SELECTIONS_OIDC_CLIENT_ID"`
	OIDCClientSecret string `env:"SELECTIONS_OIDC_CLIENT_SECRET"`
	SessionSecret    string `env:"SELECTIONS_SESSION_SECRET"`
	SessionState     string `env:"SELECTIONS_SESSION_STATE"`
	DomainName       string `env:"SELECTIONS_DOMAIN_NAME"`
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
		env.DomainName,
		env.DomainName+"/auth/callback",
		env.DomainName+"/auth/login",
		[]string{"profile", "email", "groups"},
	)

	router := gin.Default()
	//Define routes here
	router.GET("/auth/login", auth.AuthRequest)
	router.GET("/auth/callback", auth.AuthCallback)
	router.GET("/auth/logout", auth.AuthLogout)
	router.GET("/", auth.AuthWrapper())
}
