package submissions

import (
	"database/sql"
	"log"
	"net/http"
	"server/constants"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type SubmissionModal struct {
	DB *sql.DB
}

type Submission struct {
	ID        uuid.UUID
	UserId    uuid.UUID `json:"userId"`
	ProblemId uuid.UUID
	Language  string
	Code      string
	CreatedAt time.Time
	Status    string
}

func (modal *SubmissionModal) Create(submission *Submission) constants.ErrorStruct {
	query := `INSERT INTO public.submissions
	(id, user_id, problems_id, "language", code, created_at, "status")
	VALUES(uuid_generate_v4(), $1, $2, $3, $4, now(), $5) RETURNING id;`
	rows, err := modal.DB.Query(query, submission.UserId, submission.ProblemId, submission.Language, submission.Code, submission.Status)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == "submissions_problems_id_fkey" {
				return constants.ErrorStruct{
					Code:    http.StatusNotFound,
					Message: "Problem not exist",
				}
			}
		}
		log.Fatal(err)
		return constants.ErrorStruct{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	rows.Scan(&submission.ID)
	return constants.ErrorStruct{}
}

// func (modal *SubmissionModal) UserSubmissions(user *users.User) *[]Submission {
// 	query := `SELECT id, , problems_id, "language", created_at, "status"
// 	FROM submissions where created_by=$1;`

// }
// func (modal *SubmissionModal) GetSubmission(id uuid.UUID) *Submission {

// 	return submission
// }
