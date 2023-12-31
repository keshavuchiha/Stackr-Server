package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"server/error_constants"

	"github.com/google/uuid"
	pq "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserModal struct {
	DB *sql.DB
}

type User struct {
	ID           uuid.UUID
	UserName     string `json:"username,omitempty"`
	Email        string `json:"email,omitempty"`
	Password     string `json:"password"`
	RegisteredAt string
	UpdatedAt    string
}

func (userModal *UserModal) Login(user *User) error {
	query := `SELECT id, username, email, password
	FROM users WHERE users.username=$1 LIMIT 1;`
	rows, err := userModal.DB.Query(query, user.UserName)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO USER WITH GIVEN ID")
		}
		panic(err)
	}
	defer rows.Close()
	if rows.Next() {
		// user := User{}
		hashedPassword := ""
		rows.Scan(&user.ID, &user.UserName, &user.Email, &hashedPassword)
		fmt.Println(hashedPassword, user.Password)
		err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("NO USER IS FOUND")
}

func (userModal *UserModal) Register(user *User) (uuid.UUID, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, err
	}
	user.Password = string(passwordBytes)
	query := `INSERT INTO public.users
	(id, username, email, "password", registered_at, updated_at)
	VALUES(uuid_generate_v4(), $1, $2, $3, now(), now()) RETURNING id;`
	rows, err := userModal.DB.Query(query, user.UserName, user.Email, user.Password)
	if err != nil {
		// if err==sql.Err
		// if err==
		pqErr, ok := err.(*pq.Error)
		if ok {
			if pqErr.Constraint == "unique_name" {
				return uuid.UUID{}, errors.New(error_constants.UNIQUE_NAME)
			} else if pqErr.Constraint == "unique_email" {
				return uuid.UUID{}, errors.New(error_constants.UNIQUE_EMAIL)
			}
			temp, _ := json.Marshal(pqErr)
			fmt.Println(string(temp))
			fmt.Println(pqErr.Code)
			fmt.Println(pqErr)
			fmt.Printf("%+v", pqErr)
			log.Fatal(pqErr)
		}
		log.Fatalf("error in users %v", err)
	}
	id := uuid.UUID{}
	defer rows.Close()
	if rows.Next() {
		id = uuid.UUID{}
		err = rows.Scan(&id)
		fmt.Println(id)
		user.ID = id
		fmt.Println("id returned", id)
		fmt.Println(user)
	}
	return id, err
}
