package routes

import (
	"encoding/json"
	"net/http"
	"server/constants"
	"server/problems"

	"github.com/google/uuid"
)

func InsertProblem(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("user-id")
	w.Header().Add("content-type", "application/json")
	var problem problems.Problem
	problemModal := &problems.ProblemModel{
		DB: constants.DB,
	}
	problem.CreatedBy = id.(uuid.UUID)
	err := json.NewDecoder(r.Body).Decode(&problem)
	if err != nil {
		constants.ReturnError(constants.ErrorStruct{
			Code:    http.StatusBadRequest,
			Message: "Request is invalid",
		}, w)
		return
	}
	errorStruct := problemModal.Insert(&problem)
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
