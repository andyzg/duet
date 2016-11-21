package data

import (
	"encoding/json"
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
	_, err := AuthUserId(token)
	if err != nil {
		log.Printf("Error verifying token in /oauth/todoist/callback: %s", err.Error())
		http.Error(w, "Invalid Oauth2 state", http.StatusUnauthorized)
		return
	}

	code := r.FormValue("code")
	log.Printf("Retrieved Todoist code '%s'", code)

	exchangeValues := url.Values{}
	exchangeValues.Set("code", code)
	exchangeValues.Set("client_id", todoistConf.ClientID)
	exchangeValues.Set("client_secret", todoistConf.ClientSecret)
	exchangeResponse, err := http.PostForm(todoistConf.Endpoint.TokenURL, exchangeValues)
	if err != nil {
		log.Printf("Todoist code exchange failed with '%s'", err)
		http.Error(w, "Todoist code exchange failed", http.StatusUnauthorized)
		return
	}

	accessToken := AccessTokenResponse{}
	json.NewDecoder(exchangeResponse.Body).Decode(&accessToken)
	exchangeResponse.Body.Close()
	if accessToken.Error != "" {
		log.Printf("Todoist access token JSON contains error '%s'", accessToken.Error)
		http.Error(w, "Todoist code exchange failed", http.StatusUnauthorized)
		return
	}
	log.Printf("Retrieved Todoist access token '%s'", accessToken.AccessToken)

	v := url.Values{}
	v.Set("token", accessToken.AccessToken)
	v.Set("sync_token", "*")
	v.Set("resource_types", "[\"items\"]")
	response, err := http.PostForm(todoistSyncUrl, v)
	if err != nil {
		log.Printf("Error fetching data from Todoist: '%s'", err)
		http.Error(w, "Error fetching Todoist tasks", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading Todoist response: '%s'", err)
		http.Error(w, "Error reading Todoist response", http.StatusInternalServerError)
		return
	}
	log.Printf("Todoist response: '%s'", contents)
}
