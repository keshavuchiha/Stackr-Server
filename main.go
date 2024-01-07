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
	"server/routes"
	"server/secure"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTData struct {
	jwt.RegisteredClaims
	Name string `json:"name"`
}

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
		var response constants.Response
		response.Error = constants.ErrorStruct{
			Code:    http.StatusBadGateway,
			Message: "Bad Request",
		}
		res, _ = json.Marshal(&response)
		w.Write(res)
		// w.WriteHeader(http.StatusBadGateway)
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

func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		tokenString := req.Header.Get(constants.AUTHORIZATION)
		claims := &JWTData{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return ed25519.PublicKey(constants.TOKEN_PUBLIC_KEY), nil
		})
		if err != nil {
			constants.ReturnError(constants.ErrorStruct{
				Code:    http.StatusUnauthorized,
				Message: "Session has expired, Please login again",
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
			decrypted, err := secure.Decrypt(constants.PRIVATE_KEY, nameBytes)
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
	constants.DB = db
	constants.TOKEN_PUBLIC_KEY, _ = hex.DecodeString(os.Getenv("TOKEN_PUBLIC_KEY"))
	constants.TOKEN_PRIVATE_KEY, _ = hex.DecodeString(os.Getenv("TOKEN_PRIVATE_KEY"))
	constants.PRIVATE_KEY, _ = hex.DecodeString(os.Getenv("PRIVATE_KEY"))
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
	r.Post("/v1/register", routes.RegisterUser)
	r.Post("/v1/login", routes.LoginUser)
	r.Get("/v1/problems", routes.GetAllProblems)
	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Get("/v1/auth", healthcheck)
		r.Post("/v1/submission", routes.CreateSubmission)
		r.Post("/v1/problem", routes.InsertProblem)
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
