package constants

import (
	"encoding/json"
	"net/http"
)

const UNIQUE_NAME = "UNIQUE_NAME"
const UNIQUE_EMAIL = "UNIQUE_EMAIL"

type ErrorStruct struct {
	Code      int      `json:"code"`
	Message   string   `json:"message"`
	ErrorList []string `json:"errorList"`
}

func ReturnError(errorStruct ErrorStruct, w http.ResponseWriter) {
	var response Response
	response.Error = errorStruct
	responseBytes, _ := json.Marshal(&response)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(errorStruct.Code)
	w.Write(responseBytes)

}
