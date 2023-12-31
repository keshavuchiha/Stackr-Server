package main

import (
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/secure"
	"server/users"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type Claims struct {
	Name string `json:"name"`
}
type JWTData struct {
	jwt.StandardClaims
	Name string `json:"name"`
}

var TOKEN_PUBLIC_KEY []byte
var TOKEN_PRIVATE_KEY []byte
var PRIVATE_KEY []byte
var PUBLIC_KEY []byte
var db *sql.DB
var err error
var connStr string

func healthcheck(w http.ResponseWriter, req *http.Request) {

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
	http.SetCookie(w, &http.Cookie{
		Name:    "auth-token",
		Value:   tokenString,
		Expires: time.Now().Add(time.Minute * 5),
	})
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
	http.SetCookie(w, &http.Cookie{
		Name:     "auth-token",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Minute * 5),
		HttpOnly: true,
		Secure:   true,
	})
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

// func MyMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// create new context from `r` request context, and assign key `"user"`
// 		// to value of `"123"`
// 		ctx := context.WithValue(r.Context(), "user", "123")

//			// call the next handler in the chain, passing the response writer and
//			// the updated request object with the new context value.
//			//
//			// note: context.Context values are nested, so any previously set
//			// values will be accessible as well, and the new `"user"` key
//			// will be accessible from this point forward.
//			next.ServeHTTP(w, r.WithContext(ctx))
//		})
//	}

func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("auth-token")
		if err != nil {
			log.Fatal(err)
		}
		tokenString := cookie.Value
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
			fmt.Println(decrypted)
			
		} else {
			log.Fatal("error in parsing")
			fmt.Println("error cookie")
		}
		fmt.Println(cookie)
		next.ServeHTTP(w, req)
	})
}
func main() {
	DB_START()
	// fmt.Println(connStr)
	defer db.Close()
	TOKEN_PUBLIC_KEY, _ = hex.DecodeString(os.Getenv("TOKEN_PUBLIC_KEY"))
	TOKEN_PRIVATE_KEY, _ = hex.DecodeString(os.Getenv("TOKEN_PRIVATE_KEY"))
	PRIVATE_KEY, _ = hex.DecodeString(os.Getenv("PRIVATE_KEY"))

	// text := "Hello world!"
	// textHex := hex.EncodeToString([]byte(text))
	// res, err := encrypt(PRIVATE_KEY, []byte(textHex))
	// fmt.Println(res, err)
	// decrypted, err := decrypt(PRIVATE_KEY, res)
	// fmt.Println(string(decrypted), err)
	// d, _ := hex.DecodeString(string(decrypted))
	// fmt.Println(string(d))
	// PRIVATE_KEY, _ = hex.DecodeString(os.Getenv("PRIVATE_KEY"))
	// privateKey, err := ecdh.P521().GenerateKey(rand.Reader)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("ecdh private key:", hex.EncodeToString(privateKey.Bytes()))
	// fmt.Println("ecdh public key:", hex.EncodeToString(privateKey.PublicKey().Bytes()))
	fmt.Println(TOKEN_PRIVATE_KEY, TOKEN_PUBLIC_KEY)
	r := chi.NewRouter()
	r.Use(middleware.CleanPath)
	r.Use(middleware.Logger)
	r.Get("/v1/healthcheck", healthcheck)
	r.Post("/v1/problems/{id}", getProblem)
	r.Post("/v1/register", registerUser)
	r.Post("/v1/login", loginUser)

	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Get("/v1/auth", healthcheck)
	})
	// db.Exec(`delete from users;`)
	http.ListenAndServe(":8090", r)

}
