package submissions

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type SubmissionModal struct {
	DB *sql.DB
}

type Submission struct {
	ID        uuid.UUID
	UserId    uuid.UUID
	ProblemId uuid.UUID
	Language  string
	Code      string
	CreatedAt time.Time
	Status    string
}

func (modal *SubmissionModal) Create(submission *Submission) error {
	query := `INSERT INTO public.submissions
	(id, user_id, problems_id, "language", code, created_at, "status")
	VALUES(uuid_generate_v4(), $1, $2, $3, $4, now(), $5) RETURNING id;`
	rows, err := modal.DB.Query(query, submission.UserId, submission.ProblemId, submission.Language, submission.Code, submission.Status)
	if err != nil {
		return err
	}
	rows.Scan(&submission.ID)
	return nil
}
