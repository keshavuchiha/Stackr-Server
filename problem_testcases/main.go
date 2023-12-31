package problem_testcases

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type TestCasesModal struct {
	DB *sql.DB
}

type TestCase struct {
	ID        uuid.UUID
	ProblemID uuid.UUID
	Input     string
	expected  string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
}

func (testCaseModal *TestCasesModal) Add(testCase *TestCase) error {
	query := `INSERT INTO public.testcases
	(id, problem_id, "input", expected, created_by)
	VALUES(uuid_generate_v4(), $1, $2, $3, $4) RETURNING id;`
	rows, err := testCaseModal.DB.Query(query, testCase.ProblemID, testCase.Input, testCase.expected, testCase.CreatedBy)
	if err != nil {
		return err
	}
	err = rows.Scan(&testCase.ID)
	if err != nil {
		return err
	}
	return nil
}
