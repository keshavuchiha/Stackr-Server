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
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JWTData struct {
	jwt.StandardClaims
	Name string `json:"name"`
}

var TOKEN_PUBLIC_KEY []byte
var TOKEN_PRIVATE_KEY []byte
var PRIVATE_KEY []byte
var PUBLIC_KEY []byte
var db *sql.DB
var connStr string

func healthcheck(w http.ResponseWriter, req *http.Request) {
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
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	err = userModal.Login(&user)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	userIdEncrypted, err := secure.Encrypt(PRIVATE_KEY, user.ID[:])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &jwt.MapClaims{
		"name": hex.EncodeToString(userIdEncrypted),
		"exp":  time.Now().Add(time.Minute * 10).Unix(),
	})
	tokenString, err := token.SignedString(ed25519.PrivateKey(TOKEN_PRIVATE_KEY))
	w.Header().Add("content-type", "application/json")
	w.Header().Add(constants.AUTHORIZATION, tokenString)
	w.WriteHeader(http.StatusOK)
}
func registerUser(w http.ResponseWriter, req *http.Request) {
	var user users.User
	userModal := &users.UserModal{
		DB: db,
	}
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	id, err := userModal.Register(&user)
	// fmt.Println("iid", id)
	user.ID = id
	fmt.Println(user)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	fmt.Println("user id", user.ID)
	userIdEncrypted, err := secure.Encrypt(PRIVATE_KEY, user.ID[:])
	fmt.Println(err)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &jwt.MapClaims{
		"name": hex.EncodeToString(userIdEncrypted),
		"exp":  time.Now().Add(time.Minute * 10).Unix(),
	})
	tokenString, err := token.SignedString(ed25519.PrivateKey(TOKEN_PRIVATE_KEY))
	fmt.Println(tokenString)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	// claims := jwt.MapClaims{}
	claims := &JWTData{}
	token, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// token.Method.Verify()
		return ed25519.PublicKey(TOKEN_PUBLIC_KEY), nil
	})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fmt.Println(token.Claims.Valid())
	if claims, ok := token.Claims.(*JWTData); ok {
		name := claims.Name
		fmt.Printf("name: %v\n\n", name)
		nameBytes, err := hex.DecodeString(name)

		if err != nil {
			panic(err)
		}
		decrypted, err := secure.Decrypt(PRIVATE_KEY, nameBytes)
		fmt.Println(uuid.UUID(decrypted))

	} else {
		log.Fatalf("TOKEN INVALID")
	}
	fmt.Println(token)
	w.Header().Add("content-type", "application/json")
	w.Header().Add(constants.AUTHORIZATION, tokenString)
	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "auth-token",
	// 	Value:    tokenString,
	// 	Expires:  time.Now().Add(time.Minute * 5),
	// 	HttpOnly: true,
	// 	Secure:   true,
	// })
	_, err = json.Marshal(map[string]interface{}{
		"name":         user.UserName,
		"email":        user.Email,
		"registeredAt": user.RegisteredAt,
		"password":     user.Password,
	})
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	// w.Write(userBytes)
}

func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		tokenString := req.Header.Get(constants.AUTHORIZATION)
		claims := &JWTData{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// token.Method.Verify()
			return ed25519.PublicKey(TOKEN_PUBLIC_KEY), nil
		})
		if err != nil {
			log.Fatal(err)
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
	r.Get("/v1/healthcheck", healthcheck)
	r.Post("/v1/problems/{id}", getProblem)
	r.Post("/v1/register", registerUser)
	r.Post("/v1/login", loginUser)

	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Get("/v1/auth", healthcheck)
	})
	// db.Exec(`delete from users;`)
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = ":8070"
	}
	http.ListenAndServe(PORT, r)

}
