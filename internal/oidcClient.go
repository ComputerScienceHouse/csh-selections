package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	Uuid     string `json:"id"`
	Username string `json:"username"`
	Gatekeep bool   `json:"result"`
	SlackUID string `json:"slackuid"`
}

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
		fmt.Println(resp.StatusCode, resp.Status)
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

func (client *OIDCClient) GetActiveUsers() []OIDCUser {
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
		log.Println(err)
		return nil
	}
	return ret
}
