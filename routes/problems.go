package routes

import (
	"encoding/json"
	"net/http"
	"server/constants"
	"server/problems"

	"github.com/google/uuid"
)

func GetAllProblems(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	var problemFilter problems.ProblemFilter
	err := json.NewDecoder(r.Body).Decode(&problemFilter)
	if err != nil {
		constants.ReturnError(constants.ErrorStruct{
			Code:    http.StatusBadRequest,
			Message: "Request is invalid",
		}, w)
		return
	}
	problems, errorStruct := problems.GetAll(&problemFilter)
	if errorStruct.Code != 0 {
		constants.ReturnError(errorStruct, w)
		return
	}
	var response constants.Response
	response.Data = problems
	responseBytes, _ := json.Marshal(&response)
	w.Write(responseBytes)
}
func InsertProblem(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(constants.USER_ID)
	w.Header().Add("content-type", "application/json")
	var problem problems.Problem
	problem.CreatedBy = id.(uuid.UUID)
	err := json.NewDecoder(r.Body).Decode(&problem)
	if err != nil {
		constants.ReturnError(constants.ErrorStruct{
			Code:    http.StatusBadRequest,
			Message: "Request is invalid",
		}, w)
		return
	}
	errorStruct := problems.Insert(&problem)
	if errorStruct.Code != 0 {
		constants.ReturnError(errorStruct, w)
		return
	}
	w.Header().Add("content-type", "application/json")
	// w.Header().Add(constants.AUTHORIZATION, tokenString)
	var response constants.Response
	response.Error = errorStruct
	responseBytes, _ := json.Marshal(&response)
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}

// func ListProblems(w http.ResponseWriter, r http.Request) {
// 	id := r.Context().Value("user-id")
// 	w.Header().Add("content-type", "application/json")
// }
