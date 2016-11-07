package data

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

type UsernameAndPassword struct {
	Username string
	Password string
}

func ServeLogin(w rest.ResponseWriter, r *rest.Request) {
	userAndPass := UsernameAndPassword{}
	err := r.DecodeJsonPayload(&userAndPass)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(userAndPass)
}
