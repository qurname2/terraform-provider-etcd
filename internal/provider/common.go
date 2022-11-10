package provider

import (
	uuid "github.com/satori/go.uuid"
)

//uuidGenerator return random uuid nn a string format that are intended to be used as unique identifiers.
func uuidGenerator() string {
	uu := uuid.NewV4()
	return uu.String()
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
