package data

import (
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/dgrijalva/jwt-go"
)

type usernameAndPassword struct {
	Username string
	Password string
}

var tokenSecret []byte = []byte("someSecret")

func ServeLogin(w rest.ResponseWriter, r *rest.Request) {
	userAndPass := usernameAndPassword{}
	err := r.DecodeJsonPayload(&userAndPass)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO verify password
	log.Printf("Username: %s, Password: %s\n", userAndPass.Username, userAndPass.Password)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": userAndPass.Username,
	})

	tokenString, err := token.SignedString(tokenSecret)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(map[string]string{
		"token": tokenString,
	})
}
