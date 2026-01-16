package idgen

import (
	"math/rand"
	"strings"
)

func GenerateID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 6)
	for i := range id {
		id[i] = letters[rand.Intn(len(letters))]
	}
	return string(id)
}

func Name2ID(name string) string {
	return strings.ToLower(strings.Replace(strings.TrimSpace(name), " ", "_", -1)) + "_" + GenerateID()
}
