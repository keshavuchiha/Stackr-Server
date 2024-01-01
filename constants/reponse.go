package constants

type Response struct {
	Data  interface{} `json:"data"`
	Error ErrorStruct `json:"error"`
}
