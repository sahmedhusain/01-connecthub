package security

import (
	"github.com/google/uuid"
)

func GenerateToken() (uuid.UUID, error) {
	token, err := uuid.NewRandom()
	return token, err
}
