package routes

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"server/constants"
	"server/secure"
	"server/users"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func RegisterUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "application/json")
	var user users.User
	userModal := &users.UserModal{
		DB: constants.DB,
	}

	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		constants.ReturnError(constants.ErrorStruct{
			Code:    http.StatusBadRequest,
			Message: "Request is invalid",
		}, w)
		return
	}
	errorStruct := userModal.Register(&user)
	if errorStruct.Code != 0 {
		constants.ReturnError(errorStruct, w)
		return
	}
	// fmt.Println("iid", id)
	userIdEncrypted, err := secure.Encrypt(constants.PRIVATE_KEY, user.ID[:])
	fmt.Println(err)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &jwt.MapClaims{
		"name": hex.EncodeToString(userIdEncrypted),
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(ed25519.PrivateKey(constants.TOKEN_PRIVATE_KEY))
	// fmt.Println(tokenString)
	if err != nil {
		errorStruct.Code = http.StatusUnauthorized
		errorStruct.Message = err.Error()
		constants.ReturnError(errorStruct, w)
		return
	}
	var response constants.Response
	var userData users.UserData
	userData.UserName = user.UserName
	response.Data = userData
	w.Header().Add("content-type", "application/json")
	w.Header().Add(constants.AUTHORIZATION, tokenString)
	responseBytes, _ := json.Marshal(&response)
	w.WriteHeader(http.StatusCreated)
	w.Write(responseBytes)

	// w.Write(userBytes)
}

func LoginUser(w http.ResponseWriter, req *http.Request) {
	var user users.User
	userModal := &users.UserModal{
		DB: constants.DB,
	}
	json.NewDecoder(req.Body).Decode(&user)
	errorStruct := userModal.Login(&user)
	if errorStruct.Code != 0 {
		constants.ReturnError(errorStruct, w)
		return
	}
	userIdEncrypted, _ := secure.Encrypt(constants.PRIVATE_KEY, user.ID[:])
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &jwt.MapClaims{
		"name": hex.EncodeToString(userIdEncrypted),
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString(ed25519.PrivateKey(constants.TOKEN_PRIVATE_KEY))
	var response constants.Response
	var userData users.UserData
	userData.UserName = user.UserName
	response.Data = userData
	w.Header().Add("content-type", "application/json")
	w.Header().Add(constants.AUTHORIZATION, tokenString)
	responseBytes, _ := json.Marshal(&response)
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}
