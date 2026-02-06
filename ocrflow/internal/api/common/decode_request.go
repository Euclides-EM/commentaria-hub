package common

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func DecodeBody(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("failed to decode request body to type %T: %w", dst, err)
	}
	return nil
}
