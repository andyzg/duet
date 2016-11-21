package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
)

var todoistSyncUrl string = "https://todoist.com/API/v7/sync"

var todoistConf *oauth2.Config = &oauth2.Config{
	RedirectURL:  "https://api.helloduet.com/oauth/todoist/callback",
	ClientID:     os.Getenv("TODOIST_ID"),
	ClientSecret: os.Getenv("TODOIST_SECRET"),
	Scopes:       []string{"data:read"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://todoist.com/oauth/authorize",
		TokenURL: "https://todoist.com/oauth/access_token",
	},
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

type TodoistItem struct {
	Title   string `json:"content"`
	DueDate string `json:"due_date_utc"` // in format "Mon 07 Aug 2006 12:34:56 +0000"
}

func HandleTodoistLogin(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	_, err := VerifyToken(token)
	if err != nil {
		log.Printf("Error verifying token in /oauth/todist/login: %s", err.Error())
		http.Error(w, "Invalid token. URL must have token as query parameter.", http.StatusUnauthorized)
		return
	}

	url := todoistConf.AuthCodeURL(token)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleTodoistCallback(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("state")
	userId, err := AuthUserId(token)
	if err != nil {
		log.Printf("Error verifying token in /oauth/todoist/callback: %s", err.Error())
		http.Error(w, "Invalid Oauth2 state", http.StatusUnauthorized)
		return
	}

	code := r.FormValue("code")
	log.Printf("Retrieved Todoist code '%s'", code)

	v := url.Values{}
	v.Set("code", code)
	v.Set("client_id", todoistConf.ClientID)
	v.Set("client_secret", todoistConf.ClientSecret)
	response, err := http.PostForm(todoistConf.Endpoint.TokenURL, v)
	if err != nil {
		log.Printf("Todoist code exchange failed with '%s'", err)
		http.Error(w, "Todoist code exchange failed", http.StatusUnauthorized)
		return
	}

	accessToken := AccessTokenResponse{}
	json.NewDecoder(response.Body).Decode(&accessToken)
	response.Body.Close()
	if accessToken.Error != "" {
		log.Printf("Todoist access token JSON contains error '%s'", accessToken.Error)
		http.Error(w, "Todoist code exchange failed", http.StatusUnauthorized)
		return
	}
	log.Printf("Retrieved Todoist access token '%s'", accessToken.AccessToken)

	err = SyncTodoist(userId, accessToken.AccessToken)
	if err != nil {
		log.Printf("Todoist syncing failed with error '%s'", err)
		http.Error(w, "Failed to sync Todoist tasks", http.StatusUnauthorized)
		return
	}
}

func SyncTodoist(userId uint64, oauthToken string) error {
	v := url.Values{}
	v.Set("token", oauthToken)
	v.Set("sync_token", "*")
	v.Set("resource_types", "[\"items\"]")
	response, err := http.PostForm(todoistSyncUrl, v)
	if err != nil {
		log.Printf("Error making POST request to %s: '%s'", todoistSyncUrl, err)
		return fmt.Errorf("Error making request to Todoist")
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading Todoist response: '%s'", err)
		return fmt.Errorf("Error reading Todoist response")
	}
	log.Printf("Todoist response: '%s'", contents)
	return nil
}
