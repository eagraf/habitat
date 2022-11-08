package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func BindPostRequest(r *http.Request, target interface{}) error {
	if r.Method != http.MethodPost {
		return errors.New("method not POST")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading POST request body: %s", err)
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		return fmt.Errorf("error unmarshaling body into target: %s", err)
	}

	return nil
}
