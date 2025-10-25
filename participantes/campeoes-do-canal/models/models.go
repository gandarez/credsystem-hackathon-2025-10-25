package models

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func BindAndValidate(data interface{}, r *http.Request) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(data); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	// ensure there's no extra JSON after the first value
	if dec.More() {
		return fmt.Errorf("unexpected additional JSON in body")
	}

	if err := validate.Struct(data); err != nil {
		return err
	}
	return nil
}
