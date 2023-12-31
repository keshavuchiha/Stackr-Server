package main

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"time"

// 	"github.com/golang-jwt/jwt/v5"
// 	"golang.org/x/crypto/bcrypt"
// 	"google.golang.org/grpc/status"
// )

// type User struct {
// 	Name     string `json:"name"`
// 	Email    string `json:"email"`
// 	Password string `json:"password,-"`
// }

// func (u *User) Register(ctx context.Context) error {

// 	// db,err:=sql.Open("postgres", connStr)
// 	fmt.Println("Login request received for user: ", u.Name)
// 	if db == nil {
// 		panic("db is nil")
// 	}
// 	_, err := db.Exec("select 1 into users (username,email,password) values ($1,$2,$3) LIMIT 1;", u.Name, u.Email, u.Password)
// 	if err != nil {
// 		return errors.New("Query Failed")
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodES512, jwt.MapClaims{
// 		"username": u.Name,
// 		"email":    u.Email,
// 		"exp":      time.Now().Add(time.Hour * 24).Unix(),
// 	})
// 	// token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 	// 	"foo": "bar",
// 	// 	"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
// 	// })
// 	access_token, err := token.SignedString("Secret key")
// 	claims := jwt.MapClaims{}
// 	jwt.ParseWithClaims(access_token, claims)
// 	return nil
// }

// func (s *loginService) Login(ctx context.Context, in *pb.User) (*pb.LoginResponse, error) {
// 	fmt.Println("Login request received for user: ", in.Name)
// 	if db == nil {
// 		panic("db is nil")
// 	}
// 	row := db.QueryRow("select username,password from users where username=$1;", in.Name)
// 	if row == nil {
// 		return nil, status.Error(404, "User not found")
// 	}
// 	var (
// 		username string
// 		password string
// 	)
// 	err := row.Scan(&username, &password)
// 	if err != nil {
// 		return nil, status.Error(404, "User not found")
// 	}
// 	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(in.Password))
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
// 		"username": in.Name,
// 		"exp":      time.Now().Add(time.Hour * 1).Unix(),
// 	})
// 	// token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 	// 	"foo": "bar",
// 	// 	"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
// 	// })
// 	access_token, err := token.SignedString("Secret key")

// 	return &pb.LoginResponse{Message: fmt.Sprintf("Successfully logged in user with id"), Success: true, Token: access_token}, nil
// }
