package main

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/constants"
	"server/secure"
	"server/users"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type JWTData struct {
	jwt.RegisteredClaims
	Name string `json:"name"`
}

var TOKEN_PUBLIC_KEY []byte
var TOKEN_PRIVATE_KEY []byte
var PRIVATE_KEY []byte
var PUBLIC_KEY []byte
var db *sql.DB
var connStr string

func healthcheck(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Healthcheck logs")
	id := req.Context().Value("user-id")
	fmt.Println(id)
	m := map[string]string{
		"status": "UP",
	}
	res, err := json.Marshal(m)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write([]byte(res))
}

func getProblem(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	var temp struct {
		Id      string `json:"id"`
		Message string `json:"message"`
	}
	json.NewDecoder(req.Body).Decode(&temp)
	fmt.Println(temp)
	fmt.Println(id)
	temp.Id = id
	tempBytes, _ := json.Marshal(temp)
	w.Header().Add("content-type", "application/json")
	w.Write(tempBytes)
}
func loginUser(w http.ResponseWriter, req *http.Request) {
	var user users.User
	userModal := &users.UserModal{
		DB: db,
	}
	json.NewDecoder(req.Body).Decode(&user)
	errorStruct := userModal.Login(&user)
	if errorStruct.Code != 0 {
		returnError(errorStruct, w)
		return
	}
	userIdEncrypted, _ := secure.Encrypt(PRIVATE_KEY, user.ID[:])
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &jwt.MapClaims{
		"name": hex.EncodeToString(userIdEncrypted),
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString(ed25519.PrivateKey(TOKEN_PRIVATE_KEY))
	var response constants.Response
	w.Header().Add("content-type", "application/json")
	w.Header().Add(constants.AUTHORIZATION, tokenString)
	responseBytes, _ := json.Marshal(&response)
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}
func returnError(errorStruct constants.ErrorStruct, w http.ResponseWriter) {
	var response constants.Response
	response.Error = errorStruct
	responseBytes, _ := json.Marshal(&response)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(errorStruct.Code)
	w.Write(responseBytes)

}
func registerUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "application/json")
	var user users.User
	userModal := &users.UserModal{
		DB: db,
	}

	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		returnError(constants.ErrorStruct{
			Code:    http.StatusBadRequest,
			Message: "Request is invalid",
		}, w)
		return
	}
	errorStruct := userModal.Register(&user)
	if errorStruct.Code != 0 {
		returnError(errorStruct, w)
		return
	}
	// fmt.Println("iid", id)
	userIdEncrypted, err := secure.Encrypt(PRIVATE_KEY, user.ID[:])
	fmt.Println(err)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &jwt.MapClaims{
		"name": hex.EncodeToString(userIdEncrypted),
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(ed25519.PrivateKey(TOKEN_PRIVATE_KEY))
	// fmt.Println(tokenString)
	if err != nil {
		errorStruct.Code = http.StatusUnauthorized
		errorStruct.Message = err.Error()
		returnError(errorStruct, w)
		return
	}
	var response constants.Response
	w.Header().Add("content-type", "application/json")
	w.Header().Add(constants.AUTHORIZATION, tokenString)
	responseBytes, _ := json.Marshal(&response)
	w.WriteHeader(http.StatusCreated)
	w.Write(responseBytes)

	// w.Write(userBytes)
}

func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		tokenString := req.Header.Get(constants.AUTHORIZATION)
		claims := &JWTData{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return ed25519.PublicKey(TOKEN_PUBLIC_KEY), nil
		})
		if err != nil {
			returnError(constants.ErrorStruct{
				Code:    http.StatusUnauthorized,
				Message: "Session expired",
			}, w)
			return
		}
		if claims, ok := token.Claims.(*JWTData); ok {
			fmt.Println("Name", claims.Name)
			name := claims.Name
			nameBytes, err := hex.DecodeString(name)

			if err != nil {
				panic(err)
			}
			decrypted, err := secure.Decrypt(PRIVATE_KEY, nameBytes)
			decryptedId := uuid.UUID(decrypted)
			req = req.WithContext(context.WithValue(req.Context(), "user-id", decryptedId))
		} else {
			log.Fatal("error in parsing")
			fmt.Println("error cookie")
		}
		next.ServeHTTP(w, req)
	})
}
func main() {
	fmt.Println("Application starting")
	DB_START()
	// fmt.Println(connStr)
	defer db.Close()
	TOKEN_PUBLIC_KEY, _ = hex.DecodeString(os.Getenv("TOKEN_PUBLIC_KEY"))
	TOKEN_PRIVATE_KEY, _ = hex.DecodeString(os.Getenv("TOKEN_PRIVATE_KEY"))
	PRIVATE_KEY, _ = hex.DecodeString(os.Getenv("PRIVATE_KEY"))
	// fmt.Println(TOKEN_PRIVATE_KEY, TOKEN_PUBLIC_KEY)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "Authorization"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Get("/", healthcheck)
	r.Get("/v1/healthcheck", healthcheck)
	r.Post("/v1/problems/{id}", getProblem)
	r.Post("/v1/register", registerUser)
	r.Post("/v1/login", loginUser)

	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Get("/v1/auth", healthcheck)
	})

	fmt.Println("started application")
	// db.Exec(`delete from users;`)
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8070"
	}
	fmt.Println(PORT)
	http.ListenAndServe(":"+PORT, r)

}
