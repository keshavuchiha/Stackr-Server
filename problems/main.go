package problems

import (
	"database/sql"
	"server/constants"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ProblemModel struct {
	DB *sql.DB
}
type Problem struct {
	ID          uuid.UUID
	Title       string `json:"title"`
	Description string `json:"description"`
	Score       int    `json:"score"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
}

type Status string

const (
	ALL    Status = ""
	TODO   Status = "todo"
	SOLVED Status = "solved"
	TRIED  Status = "tried"
)

type ProblemFilter struct {
	Status       Status
	Tags         []string
	Page         int
	Offset       int
	ProblemTitle string
}

// TODO: Complete it
// func (problemModal *ProblemModel) GetAll(problemFilter *ProblemFilter) {
// 	if problemFilter.Status == ALL {
// 		query := `Select * from problems as p , problem_tags as pt
// 		where problems.id=problem_tags.problem_id and pt.tag  in [problemFilter.tags]
// 		ans lower(p.title) like %problemFilter% limit offset offset (problemFilter.page-1)*offset`
// 	}
// }
// func (problemModal *ProblemModel) Get(id uuid.UUID) {
// 	query := `select * from problems where id=$1`
// }

func (probelmModel *ProblemModel) Insert(problem *Problem) constants.ErrorStruct {
	query := `INSERT INTO public.problems
	(id, title, description, score, created_at, updated_at, created_by)
	VALUES(uuid_generate_v4(), $1, $2,$3, now(), now(), $4);`
	_, err := probelmModel.DB.Exec(query, problem.Title, problem.Description, problem.Score, problem.CreatedBy)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			if pqErr.Constraint == "unique_title" {
				return constants.ErrorStruct{
					Code:    409,
					Message: constants.UNIQUE_NAME,
				}
			}
			return constants.ErrorStruct{
				Code:    400,
				Message: pqErr.Message,
			}
		}
		return constants.ErrorStruct{
			Code:    500,
			Message: err.Error(),
		}
	}
	return constants.ErrorStruct{}
}
