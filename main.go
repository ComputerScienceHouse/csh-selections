package main

import (
	"html/template"
	"net/http"
	"strings"

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
	router.FuncMap = template.FuncMap{"len": Length, "appScore": GetApplicationScore}
	router.LoadHTMLGlob("templates/*")
	router.StaticFS("/static", http.Dir("static"))
	//Define routes here
	router.GET("/auth/login", auth.AuthRequest)
	router.GET("/auth/callback", auth.AuthCallback)
	router.GET("/auth/logout", auth.AuthLogout)
	router.GET("/", auth.AuthWrapper(HandleSessionHomePage))
	router.GET("/admin/debug", auth.AuthWrapper(HandleDebugPage))
	router.POST("/admin/debug", auth.AuthWrapper(HandleDebugPost))
	router.POST("/application/upload", auth.AuthWrapper(HandleApplicationFileUpload))
	router.GET("/application/manage", auth.AuthWrapper(HandleApplicationManagementPage))
	router.GET("/application/export", auth.AuthWrapper(HandleApplicationExport))
	router.GET("/application/:id", auth.AuthWrapper(HandleApplicationGet))
	router.POST("/application/:id/rate", auth.AuthWrapper(HandleApplicationRating))
	router.DELETE("/application/:id", auth.AuthWrapper(HandleApplicationDelete))
	router.POST("/team/:id/assign", auth.AuthWrapper(HandleTeamApplicationAssignment))
	router.GET("/session/manage", auth.AuthWrapper(HandleSessionManagementPage))
	router.POST("/session/manage", auth.AuthWrapper(HandleSessionChanging))
	router.GET("/session/allMembers", auth.AuthWrapper(HandleSessionEligibleMemberList))
	router.POST("/session/addMembers", auth.AuthWrapper(HandleSessionAddMembers))
	router.POST("/session/removeMember", auth.AuthWrapper(HandleSessionRemoveMember))
	router.POST("/session/createTeams", auth.AuthWrapper(HandleTeamCreation))
	router.GET("/admin/debug/error", HandleDebugError)

	address := ":8080"
	if strings.Contains(env.BaseUri, "localhost") {
		address = "localhost" + address
	}
	router.Run(address)

}
