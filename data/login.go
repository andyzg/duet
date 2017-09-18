package data

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
        "os"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/dgrijalva/jwt-go"

	"golang.org/x/crypto/bcrypt"
)

type usernameAndPassword struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type signupInfo struct {
	Username string `json:"username"`
	Id       uint64 `json:"id"`
}

type DuetClaims struct {
	jwt.StandardClaims
}

var tokenSecret []byte = []byte(os.Getenv("JWT_SECRET"))

var bcryptCost int = 10

func ServeCreateUser(db Database) func(rest.ResponseWriter, *rest.Request) {
	return func(w rest.ResponseWriter, r *rest.Request) {
		userAndPass := usernameAndPassword{}
		err := r.DecodeJsonPayload(&userAndPass)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := db.CreateUser(userAndPass.Username, userAndPass.Password)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteJson(signupInfo{
			Username: user.Username,
			Id:       user.Id,
		})
	}
}

func ServeLogin(db Database) func(rest.ResponseWriter, *rest.Request) {
	return func(w rest.ResponseWriter, r *rest.Request) {
		userAndPass := usernameAndPassword{}
		err := r.DecodeJsonPayload(&userAndPass)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tokenString, err := Login(db, userAndPass.Username, userAndPass.Password)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusUnauthorized)
		}

		w.WriteJson(map[string]string{
			"token": tokenString,
		})
	}
}

func Login(db Database, username string, password string) (string, error) {
	user, err := db.GetUserByUsername(username)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
	if err != nil {
		return "", err
	}

	// TODO don't log password
	log.Printf("Username: %s, Password: %s\n", username, password)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, DuetClaims{
		jwt.StandardClaims{
			Subject:  strconv.FormatUint(user.Id, 10),
			Issuer:   "Duet",
			Audience: "https://api.helloduet.com",
		},
	})

	tokenString, err := token.SignedString(tokenSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ServeVerifyToken(db Database) func(rest.ResponseWriter, *rest.Request) {
	return func(w rest.ResponseWriter, r *rest.Request) {
		token, err := GetBearerToken(r.Request)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := VerifyToken(token)
		if err != nil {
			log.Printf("Error verifying token: %s", err.Error())
			rest.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		w.WriteJson(claims)
	}
}

func VerifyToken(tokenString string) (*DuetClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &DuetClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return tokenSecret, nil
	})

	// log.Printf("Verifying token %s\n", tokenString)

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*DuetClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("Token could not be parsed")
	}
}

func GetBearerToken(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")
	if !strings.HasPrefix(authorization, "Bearer ") {
		log.Printf("Invalid authorization header: \"%s\"", authorization)
		return "", fmt.Errorf("Invalid authentication method")
	}
	return strings.TrimPrefix(authorization, "Bearer "), nil
}

func AuthUserId(tokenString string) (uint64, error) {
	claims, err := VerifyToken(tokenString)
	if err != nil {
		return 0, err
	}
	userId, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}
	return userId, nil
}
