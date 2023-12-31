package contests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"server/problems"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ContestModal struct {
	DB *sql.DB
}

type Contest struct {
	ID        uuid.UUID
	Title     string
	StartTime time.Time
	EndTime   time.Time
	CreatedBy uuid.UUID
}

func (contestModal *ContestModal) Insert(contest *Contest) (*Contest, error) {
	query := `INSERT INTO public.contests
	(id, title, start_time, end_time, created_by)
	VALUES(uuid_generate_v4(), $1, $2, $3, $4) RETURNING id;`
	rows, err := contestModal.DB.Query(query, contest.Title, contest.StartTime, contest.EndTime, contest.CreatedBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rows.Scan(&contest.ID)
	return contest, nil
}

func (contestModal *ContestModal) AddProblem(conest *Contest, problem *problems.Problem) error {
	query := `INSERT INTO public.contest_problems
	(contest_id, problem_id)
	VALUES($1, $2, $3);`
	_, err := contestModal.DB.Exec(query, conest.ID, problem.ID)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			temp, _ := json.Marshal(pqErr)
			fmt.Println(temp)
			// log.Fatalf("ERROR IN ADD PROBLEM", pqErr)
			problemModal := problems.ProblemModel{
				DB: contestModal.DB,
			}
			problemModal.Insert(problem)
		}
		return err
	}
	return nil
}
