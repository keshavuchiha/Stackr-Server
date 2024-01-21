package contests

import (
	"encoding/json"
	"fmt"
	"server/constants"
	"server/problems"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Contest struct {
	ID        uuid.UUID
	Title     string
	StartTime time.Time
	EndTime   time.Time
	CreatedBy uuid.UUID
}

func Insert(contest *Contest) (*Contest, error) {
	query := `INSERT INTO public.contests
	(id, title, start_time, end_time, created_by)
	VALUES(uuid_generate_v4(), $1, $2, $3, $4) RETURNING id;`
	rows, err := constants.DB.Query(query, contest.Title, contest.StartTime, contest.EndTime, contest.CreatedBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rows.Scan(&contest.ID)
	return contest, nil
}

func AddProblem(conest *Contest, problem *problems.Problem) error {
	query := `INSERT INTO public.contest_problems
	(contest_id, problem_id)
	VALUES($1, $2, $3);`
	_, err := constants.DB.Exec(query, conest.ID, problem.ID)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			temp, _ := json.Marshal(pqErr)
			fmt.Println(temp)
			// log.Fatalf("ERROR IN ADD PROBLEM", pqErr)
			problems.Insert(problem)
		}
		return err
	}
	return nil
}
