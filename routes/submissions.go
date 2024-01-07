package routes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"server/constants"
	"server/problems"
	"server/submissions"
	"time"

	"github.com/google/uuid"
)

func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// func
func CreateSubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	submissionModal := submissions.SubmissionModal{
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
	errorStruct := submissionModal.Create(&submission)
	if errorStruct.Code != 0 {
		constants.ReturnError(errorStruct, w)
		return
	}
	var response constants.Response
	// submission.ProblemId
	//TODO:check for problem of the day
	//TODO: check for streak
	rows, _ := submissionModal.DB.Query(`select problem_id from problem_of_the_day where day=CURRENT_DATE;`)
	potd := uuid.UUID{}
	if rows.Next() {
		rows.Scan(&potd)
	}
	if potd == submission.ProblemId {
		// ctx := context.Background()
		tx, err := submissionModal.DB.Begin()
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Rollback()
		var t time.Time
		row := tx.QueryRow(`select current_streak,max_streak,last_submitted,CURRENT_TIMESTAMP from streaks where user_id=$1`, submission.UserId)

		streak := problems.Streak{}
		err = row.Scan(&streak.CurrentStreak, &streak.MaxStreak, &streak.LastSubmitted, &t)
		if err == sql.ErrNoRows {
			streak.CurrentStreak = 1
			streak.MaxStreak = 1
		} else {
			if DateEqual(streak.LastSubmitted.Add(time.Hour*24), t) {
				streak.CurrentStreak++
				streak.MaxStreak = max(streak.CurrentStreak, streak.MaxStreak)
			} else {
				streak.CurrentStreak = 1
			}
		}

		streak.LastSubmitted = t

		tx.QueryRow(`update streaks set curent_streak = $1,max_streak=$2,last_submitted=$3 where user_id=$4;`,
			streak.CurrentStreak, streak.MaxStreak, streak.LastSubmitted, submission.UserId)
		tx.Commit()
		// submissionModal.DB
		// submissionModal.DB.Query(`select user_id,current_streak,max_streak,leas from streaks where user_id=$1`)
		response.Data = problems.Streak{}
	}
	responseBytes, _ := json.Marshal(&response)
	w.Write(responseBytes)
}
