package problems

import (
	"database/sql"
	"fmt"
	"log"
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
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Score       int       `json:"score"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Companies   []string  `json:"companies,omitempty"`
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
type AllProblems struct {
	Problem_Id   uuid.UUID
	Title        string
	Description  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    uuid.UUID
	Score        int
	Acceptance   float64
	CompanyNames []string
	Tags         []string
}
type Streak struct {
	CurrentStreak   int  `json:"current_streak"`
	MaxStreak       int  `json:"max_streak"`
	ProblemOfTheDay bool `json:"problem_of_the_day"`
	LastSubmitted   time.Time
}

func (problemModal *ProblemModel) CheckForProblemOfTheDay(tx *sql.Tx) bool {
	row := tx.QueryRow(`SELECT problem_id
		FROM public.problem_of_the_day where "day"=CURRENT_DATE;`)
	err := row.Scan(&constants.ProblemOfTheDay)
	fmt.Println("Problem of the day", constants.ProblemOfTheDay)
	return err != sql.ErrNoRows
}

func (problemModal *ProblemModel) AddProblemOfTheDay(tx *sql.Tx) {
	rows, err := tx.Query(`select p.id from problems p where p.id not in (SELECT problem_id
		FROM public.problem_of_the_day) order by random() limit 1;`)
	// var problemIDs uuid.UUIDs
	if err != nil {
		log.Fatal(err)
	}
	var id uuid.UUID
	for rows.Next() {
		id = uuid.UUID{}
		rows.Scan(&id)
	}
	fmt.Println(id)
	_, err = tx.Exec(`INSERT INTO public.problem_of_the_day
		("day", problem_id)
		VALUES(CURRENT_DATE, $1);`, id)
	if err != nil {
		pqerr, ok := err.(*pq.Error)
		if ok {
			log.Fatal(pqerr)
		}
		log.Fatal(err)
	}
}
func (problemModal *ProblemModel) GetAll(problemFilter *ProblemFilter) (*[]AllProblems, constants.ErrorStruct) {
	query := `with problem_submissions as (select p.id,
		(sum(case when s."status"='Accepted' then 1 else 0 end)*1.0)/
		greatest(sum(case when s."status"!='Compiler Error' then 1 else 0 end),1) as acceptance
		from problems p left join submissions s on s.problems_id = p.id
		 group by p.id),
		 problem_tags_agg as (select p.id,array_remove(array_agg(distinct pt.tag),null) as tags from problems p
		 left join problem_tags pt on pt.problem_id  = p.id group by p.id),
		 companies_agg as (select p.id,array_remove(array_agg(distinct c.username),null) as company_names from problems p left join company_problems cp on p.id =cp.problem_id
		 left join users c on cp.company_id = c.id group by p.id)

		 select p.id as problem_id,title,description,created_at,updated_at,created_by,score,acceptance,company_names,tags from problems p
		join problem_submissions ps on ps.id=p.id
		join problem_tags_agg as pta on pta.id=p.id
		join companies_agg as ca on ca.id=p.id;`
	rows, err := problemModal.DB.Query(query)
	if err != nil {
		return nil, constants.ErrorStruct{
			Code:    400,
			Message: "Error in recieving rows",
		}
	}

	var result []AllProblems
	for rows.Next() {
		var problem AllProblems
		fmt.Println(rows.Columns())
		rows.Scan(&problem.Problem_Id, &problem.Title, &problem.Description,
			&problem.CreatedAt, &problem.UpdatedAt, &problem.CreatedBy,
			&problem.Score, &problem.Acceptance, pq.Array(&problem.CompanyNames), pq.Array(&problem.Tags))
		result = append(result, problem)

	}
	return &result, constants.ErrorStruct{}

}

func (problemModal *ProblemModel) Get(id uuid.UUID) *Problem {
	fmt.Println(id)
	row := problemModal.DB.QueryRow(`SELECT id, title, description, created_at, 
	updated_at, created_by, score
	FROM public.problems where id=$1 limit 1;`, id)
	var problem Problem
	row.Scan(&problem.ID, &problem.Title, &problem.Description,
		&problem.CreatedAt, &problem.UpdatedAt, &problem.CreatedBy, &problem.Score)
	return &problem
}

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
