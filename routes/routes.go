package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/constants"
	"server/problems"

	"github.com/google/uuid"
)

type Home struct {
	Companies       []string          `json:"companies"`
	ProblemOfTheDay *problems.Problem `json:"potd"`
}

func getCompanies(limit int) []string {
	limitString := "all"
	if limit > 0 {
		limitString = fmt.Sprint(limit)
	}
	rows, err := constants.DB.Query(`select u.username from companies c join users u on c.id =u.id 
	left join company_problems cp on cp.company_id = c.id group by u.username order by count(*) desc limit $1;`, limitString)
	if err != nil {
		log.Fatal(err)
	}
	var result []string
	for rows.Next() {
		company := ""
		rows.Scan(&company)
		result = append(result, company)
	}
	return result
}
func GetCompaniesByProblemId(id uuid.UUID) []string {
	rows, _ := constants.DB.Query(`select distinct username from company_problems cp 
join users u on cp.company_id = u.id where cp.problem_id =$1 ;`, id)
	var result []string
	for rows.Next() {
		var companyName string
		rows.Scan(&companyName)
		result = append(result, companyName)
	}
	return result
}

func GetHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	var response constants.Response
	// constants.DB.Exec()
	problem := problems.Get(constants.ProblemOfTheDay)
	problem.Companies = GetCompaniesByProblemId(problem.ID)
	response.Data = Home{
		Companies:       getCompanies(8),
		ProblemOfTheDay: problem,
	}
	responseBytes, _ := json.Marshal(&response)

	w.Write(responseBytes)
}
