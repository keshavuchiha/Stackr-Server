package constants

const UNIQUE_NAME = "UNIQUE_NAME"
const UNIQUE_EMAIL = "UNIQUE_EMAIL"

type ErrorStruct struct {
	Code      int      `json:"code"`
	Message   string   `json:"message"`
	ErrorList []string `json:"errorList"`
}
