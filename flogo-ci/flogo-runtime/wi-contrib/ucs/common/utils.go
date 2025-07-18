package common

import (
	"crypto/rand"
	"io"
	"time"

	"github.com/oklog/ulid"
)

func GetUniqueId() (string, error) {
	entropy, err := generateRandomBytes(10)
	if err != nil {
		return "", err
	}
	uuid, err := ulid.New(ulid.Timestamp(time.Now().UTC()), entropy)
	err = NewErrorWithStack(err)

	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}

func generateRandomBytes(n int) (io.Reader, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	err = NewErrorWithStack(err)

	if err != nil {
		return nil, err
	}
	return rand.Reader, nil
}
