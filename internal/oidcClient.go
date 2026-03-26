package internal

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	cshAuth "github.com/computersciencehouse/csh-auth"
)

type OIDCClient struct {
	oidcClientId     string
	oidcClientSecret string

	accessToken  string
	providerBase string
	quit         chan struct{}
}

type OIDCUser struct {
	Uuid      string   `json:"id"`
	Username  string   `json:"username"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	FullName  string   `json:"fullName"`
	Groups    []string `json:"groups"`
	Gatekeep  bool     `json:"result"`
	SlackUID  string   `json:"slackuid"`
}

var OIDCUserByUsername = make(map[string]*OIDCUser)
var OIDCUserByUuid = make(map[string]*OIDCUser)

func (client *OIDCClient) setupOidcClient(oidcClientId, oidcClientSecret string) {
	client.oidcClientId = oidcClientId
	client.oidcClientSecret = oidcClientSecret
	parse, err := url.Parse(cshAuth.ProviderURI)
	if err != nil {
		log.Println(err)
		return
	}
	client.providerBase = parse.Scheme + "://" + parse.Host
	exp := client.getAccessToken()
	ticker := time.NewTicker(time.Duration(exp) * time.Second)
	// this will async get the token
	go func() {
		for {
			select {
			case <-ticker.C:
				exp = client.getAccessToken()
				ticker.Reset(time.Duration(exp) * time.Second)
			case <-client.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (client *OIDCClient) getAccessToken() int {
	htclient := http.DefaultClient
	//request body
	authData := url.Values{}
	authData.Set("client_id", client.oidcClientId)
	authData.Set("client_secret", client.oidcClientSecret)
	authData.Set("grant_type", "client_credentials")
	resp, err := htclient.PostForm(cshAuth.ProviderURI+"/protocol/openid-connect/token", authData)
	if err != nil {
		log.Println(err)
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println(resp.StatusCode, resp.Status)
		log.Println("The OIDC credentials may not be setup for client authorization")
		return 0
	}
	respData := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Println(err)
		return 0
	}
	if respData["error"] != nil {
		log.Println(respData)
		return 0
	}
	client.accessToken = respData["access_token"].(string)
	return int(respData["expires_in"].(float64))
}

func (user OIDCUser) IsEboard() bool {
	if len(user.Groups) == 0 {
		oidcClient.GetUserInfo(user.Username)
	}
	return slices.Contains(user.Groups, "eboard")
}

func (client *OIDCClient) GetActiveUsers() []OIDCUser {
	if members, b := goCache.Get("activeMembers"); b {
		return members.([]OIDCUser)
	}
	htclient := &http.Client{}
	//active
	req, err := http.NewRequest("GET", client.providerBase+"/auth/admin/realms/csh/groups/a97a191e-5668-43f5-bc0c-6eefc2b958a7/members", nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	req.Header.Add("Authorization", "Bearer "+client.accessToken)
	resp, err := htclient.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()
	ret := make([]OIDCUser, 0)
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		log.Println("GetActiveUsers", err)
		return nil
	}
	goCache.SetDefault("activeMembers", ret)
	return ret
}

func (client *OIDCClient) GetUserInfo(username string) *OIDCUser {
	if user, ok := OIDCUserByUsername[username]; ok {
		return user
	}
	user := &OIDCUser{Username: username}
	htclient := &http.Client{}
	arg := ""
	if len(user.Uuid) == 0 {
		arg = "?exact=true&username=" + user.Username
	}
	req, err := http.NewRequest("GET", client.providerBase+"/auth/admin/realms/csh/users/"+user.Uuid+arg, nil)
	// also "users/{user-id}/groups"
	if err != nil {
		log.Println(err)
		return user
	}
	req.Header.Add("Authorization", "Bearer "+client.accessToken)
	resp, err := htclient.Do(req)
	if err != nil {
		log.Println(err)
		return user
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(b), "error") {
		log.Println(string(b))
		return user
	}
	if len(arg) > 0 {
		userData := make([]map[string]any, 0)
		err = json.Unmarshal(b, &userData)
		// userdata attributes are a KV pair of string:[]any, this casts attributes, finds the specific attribute, casts it to a list of any, and then pulls the first field since there will only ever be one
		userAttributes := userData[0]["attributes"].(map[string]any)
		user.Uuid = userData[0]["id"].(string)
		user.FirstName = userData[0]["firstName"].(string)
		user.LastName = userData[0]["lastName"].(string)
		user.FullName = user.FirstName + " " + user.LastName
		if slackIDRaw, exists := userAttributes["slackuid"]; exists {
			user.SlackUID = slackIDRaw.([]any)[0].(string)
		} else {
			log.Println("User " + user.Username + " does not have a SlackUID.")
		}
	} else {
		err = json.Unmarshal(b, &user)
	}
	if err != nil {
		log.Println(err)
	}
	OIDCUserByUuid[user.Uuid] = user
	OIDCUserByUsername[user.Username] = user
	req, err = http.NewRequest("GET", client.providerBase+"/auth/admin/realms/csh/users/"+user.Uuid+"/groups", nil)
	if err != nil {
		log.Println(err)
		return user
	}
	req.Header.Add("Authorization", "Bearer "+client.accessToken)
	resp, err = htclient.Do(req)
	if err != nil {
		log.Println(err)
		return user
	}
	defer resp.Body.Close()
	b, _ = io.ReadAll(resp.Body)
	if strings.Contains(string(b), "error") {
		log.Println(string(b))
		return user
	}
	var groups []map[string]string
	err = json.Unmarshal(b, &groups)
	if err != nil {
		log.Println(err)
		return user
	}
	for _, group := range groups {
		user.Groups = append(user.Groups, group["name"])
	}
	return user
}
