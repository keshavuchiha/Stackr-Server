package users

import (
	"database/sql"
	"net/http"
	"server/constants"
	"server/validations"

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

func (userModal *UserModal) Login(user *User) constants.ErrorStruct {
	query := `SELECT id, username, email, password
	FROM users WHERE users.email=$1 LIMIT 1;`
	rows, err := userModal.DB.Query(query, user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return constants.ErrorStruct{
				Code:    http.StatusNotFound,
				Message: "No User with given Email",
			}
		}
		return constants.ErrorStruct{
			Code:    http.StatusInternalServerError,
			Message: "Internal Server Error",
		}
	}
	defer rows.Close()
	if rows.Next() {
		// user := User{}
		hashedPassword := ""
		rows.Scan(&user.ID, &user.UserName, &user.Email, &hashedPassword)
		err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
		if err != nil {
			return constants.ErrorStruct{
				Code:    http.StatusUnauthorized,
				Message: "Password is incorrect",
			}
		} else {
			return constants.ErrorStruct{}
		}

	}
	return constants.ErrorStruct{
		Code:    http.StatusNotFound,
		Message: "User Not Found",
	}
}

func (userModal *UserModal) Register(user *User) constants.ErrorStruct {

	var errorStruct constants.ErrorStruct
	if len(user.UserName) < 5 {
		errorStruct.ErrorList = append(errorStruct.ErrorList, "Name too short")
	} else if len(user.UserName) > 50 {
		errorStruct.ErrorList = append(errorStruct.ErrorList, "Name too long")
	}
	if !validations.IsEmailValid(user.Email) {
		errorStruct.ErrorList = append(errorStruct.ErrorList, "Email is invalid")
	}
	if !validations.IsPasswordValid(user.Password) {
		errorStruct.ErrorList = append(errorStruct.ErrorList, "Password is invalid")
	}
	if len(errorStruct.ErrorList) > 0 {
		errorStruct.Code = http.StatusBadRequest
		return errorStruct
	}
	passwordBytes, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	// if err != nil {
	// 	return uuid.UUID{}, err
	// }
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
				return constants.ErrorStruct{
					Code:    409,
					Message: constants.UNIQUE_NAME,
				}
			} else if pqErr.Constraint == "unique_email" {
				return constants.ErrorStruct{
					Code:    409,
					Message: constants.UNIQUE_EMAIL,
				}
			}
			return constants.ErrorStruct{
				Code:    400,
				Message: pqErr.Message,
			}
		}
		return constants.ErrorStruct{
			Code:    500,
			Message: err.Error(),
		}
	}
	// id := uuid.UUID{}
	defer rows.Close()
	if rows.Next() {
		id := uuid.UUID{}
		_ = rows.Scan(&id)
		// fmt.Println(id)
		user.ID = id
		// fmt.Println("id returned", id)
		// fmt.Println(user)
	}
	return constants.ErrorStruct{}
}
