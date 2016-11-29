package data

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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
	Checked int    `json:checked`        // 1 is true, 0 is false
}

type TodoistSync struct {
	SyncToken string        `json:"sync_token"`
	Items     []TodoistItem `json:"items"`
}

func HandleTodoistLogin(db Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		userId, err := AuthUserId(token)
		if err != nil {
			log.Printf("Error verifying token in /oauth/todist/login: %s", err.Error())
			http.Error(w, "Invalid token. URL must have token as query parameter.", http.StatusUnauthorized)
			return
		}

		log.Printf("Redirecting to Todoist for user %d", userId)
		url := todoistConf.AuthCodeURL(token)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})
}

func HandleTodoistCallback(db Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		err = SyncTodoist(db, userId, accessToken.AccessToken)
		if err != nil {
			log.Printf("Todoist syncing failed with error '%s'", err)
			http.Error(w, "Failed to sync Todoist tasks", http.StatusUnauthorized)
			return
		}
	})
}

func SyncTodoist(db Database, userId uint64, oauthToken string) error {
	v := url.Values{}
	v.Set("token", oauthToken)
	v.Set("sync_token", "*")
	v.Set("resource_types", "[\"items\"]")
	response, err := http.PostForm(todoistSyncUrl, v)
	if err != nil {
		log.Printf("Error making POST request to %s: '%s'", todoistSyncUrl, err)
		return fmt.Errorf("Error making request to Todoist")
	}

	sync := TodoistSync{}
	err = json.NewDecoder(response.Body).Decode(&sync)
	response.Body.Close()
	if err != nil {
		log.Printf("Error parsing Todoist response: '%s'", err)
		return fmt.Errorf("Error parsing Todoist response")
	}

	for _, item := range sync.Items {
		var endDate *time.Time
		if item.DueDate != "" {
			const longForm = "Mon 02 Jan 2006 15:04:05 +0000"
			t, err := time.Parse(longForm, item.DueDate)
			if err != nil {
				log.Printf("Error parsing due date '%s': '%s'", item.DueDate)
			} else {
				endDate = &t
			}
		}
		task := Task{
			Kind:    TaskEnum,
			Title:   item.Title,
			Done:    item.Checked == 1,
			EndDate: endDate,
		}
		err = db.AddTask(&task, userId)
		if err != nil {
			log.Printf("Error adding task: '%s'", err)
		} else {
			log.Printf("Added task %s", task)
		}
	}

	return nil
}
