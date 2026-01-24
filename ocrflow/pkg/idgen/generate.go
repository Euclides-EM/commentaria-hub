package idgen

import (
	"fmt"
	"math/rand"
	"strings"
)

func GenerateID(prefix string) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 6)
	for i := range id {
		id[i] = letters[rand.Intn(len(letters))]
	}

	if prefix == "" {
		return string(id)
	}
	return fmt.Sprintf("%s_%s", prefix, string(id))
}

func Name2ID(prefix, name string) string {
	id := fmt.Sprintf("%s_%s", strings.ToLower(strings.Replace(strings.TrimSpace(name), " ", "_", -1)), GenerateID(""))
	if prefix != "" {
		id = fmt.Sprintf("%s_%s", prefix, id)
	}
	return id
}
