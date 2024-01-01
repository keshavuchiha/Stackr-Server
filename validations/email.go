package validations

import (
	"net/mail"
)

func IsEmailValid(e string) bool {
	_, err := mail.ParseAddress(e)
	return err == nil
}
