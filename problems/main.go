package problems

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type ProblemModel struct {
	DB *sql.DB
}
type Problem struct {
	ID          uuid.UUID
	Title       string
	Description string
	Score       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
}

func (probelmModel *ProblemModel) Insert(problem *Problem) error {
	query := `INSERT INTO public.problems
	(id, title, description, score, created_at, updated_at, created_by)
	VALUES(uuid_generate_v4(), $1, $2, now(), now(), $3);`
	_, err := probelmModel.DB.Exec(query, problem.Title, problem.Description,problem.Score, problem.CreatedBy)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
