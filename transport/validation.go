package transport

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

const (
	ParamValidationError = "params_validation"
	BodyValidationError  = "body_validation"
	QueryValidationError = "query_validation"
)

type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{validate: validator.New()}
}

func bindAndValidate[T any](validate *validator.Validate, data []byte, errorType string) (*T, error) {
	var target T
	if err := json.Unmarshal(data, &target); err != nil {
		return nil, BadRequest(err.Error(), errorType, nil)
	}
	if err := validate.Struct(target); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			return nil, ValidationError(errorType, verr)
		}
		return nil, err
	}
	return &target, nil
}

func BindAndValidateBody[T any](v *Validator, r *http.Request) (*T, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, BadRequest("request body is required", BodyValidationError, nil)
	}
	return bindAndValidate[T](v.validate, body, BodyValidationError)
}

func BindAndValidateParams[T any](v *Validator, r *http.Request) (*T, error) {
	vars := mux.Vars(r)
	data, err := json.Marshal(vars)
	if err != nil {
		return nil, err
	}
	return bindAndValidate[T](v.validate, data, ParamValidationError)
}

func BindAndValidateQuery[T any](v *Validator, r *http.Request) (*T, error) {
	q := make(map[string]any)
	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			q[key] = values[0]
		} else {
			q[key] = values
		}
	}
	data, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	return bindAndValidate[T](v.validate, data, QueryValidationError)
}
