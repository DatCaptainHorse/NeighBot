package utilities

import "github.com/google/uuid"

func NewContextID() string {
	return uuid.NewString()
}
