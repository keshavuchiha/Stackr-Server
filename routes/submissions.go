package routes

import (
	"encoding/json"
	"net/http"
	"server/constants"
	"server/submissions"

	"github.com/google/uuid"
)

func CreateSubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	submissonModal := submissions.SubmissionModal{
		DB: constants.DB,
	}
	var body struct {
		ProblemId string `json:"problemId"`
		Language  string `json:"language"`
		Code      string `json:"code"`
	}
	userId := r.Context().Value("user-id")
	id := userId.(uuid.UUID)

	var submission submissions.Submission
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.Code == "" || body.Language == "" {
		constants.ReturnError(constants.ErrorStruct{
			Code:    http.StatusBadRequest,
			Message: "Error in parsing json reposne",
		}, w)
		return
	}
	defer r.Body.Close()
	submission.UserId = id
	submission.ProblemId, err = uuid.Parse(body.ProblemId)
	if err != nil {
		constants.ReturnError(constants.ErrorStruct{
			Code:    http.StatusBadRequest,
			Message: "Invalid Problem Id",
		}, w)
		return
	}
	submission.Code = body.Code
	submission.Language = body.Language
	submission.Status = "Submitted"
	errorStruct := submissonModal.Create(&submission)
	if errorStruct.Code != 0 {
		constants.ReturnError(errorStruct, w)
		return
	}
	var response constants.Response
	responseBytes, _ := json.Marshal(&response)
	w.Write(responseBytes)
}
