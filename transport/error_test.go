package transport

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestValidationError(t *testing.T) {
	validate := validator.New()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"gt=0"`
	}

	ts := TestStruct{
		Name:  "",
		Email: "invalid-email",
		Age:   -5,
	}

	err := validate.Struct(ts)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	verrs := err.(validator.ValidationErrors)
	appErr := ValidationError("test_code", verrs)

	if appErr.Type != BadRequestError {
		t.Errorf("Expected status %s, got %s", BadRequestError, appErr.Type)
	}

	// Check Meta for structured errors
	metaErrors, ok := appErr.Meta["fields"].([]FieldError)
	if !ok {
		t.Fatal("Expected 'fields' in Meta to be []FieldError")
	}

	if len(metaErrors) != 3 {
		t.Errorf("Expected 3 field errors, got %d", len(metaErrors))
	}

	fieldMap := make(map[string]FieldError)
	for _, fe := range metaErrors {
		fieldMap[fe.Field] = fe
	}

	if fe, ok := fieldMap["Name"]; !ok || fe.Tag != "required" {
		t.Errorf("Expected 'required' error for Name, got %+v", fe)
	}
	if fe, ok := fieldMap["Email"]; !ok || fe.Tag != "email" {
		t.Errorf("Expected 'email' error for Email, got %+v", fe)
	}
	if fe, ok := fieldMap["Age"]; !ok || fe.Tag != "gt" {
		t.Errorf("Expected 'gt' error for Age, got %+v", fe)
	}
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	appErr := BadRequest("bad request message", "BAD_CODE", map[string]any{"key": "value"})

	err := Error(w, appErr)
	if err != nil {
		t.Fatalf("Error() returned error: %v", err)
	}

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var respBody map[string]any
	if err := json.NewDecoder(w.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if respBody["error"] != "bad_request" {
		t.Errorf("Expected error type 'bad_request', got %v", respBody["error"])
	}
	if respBody["code"] != "BAD_CODE" {
		t.Errorf("Expected code 'BAD_CODE', got %v", respBody["code"])
	}
}
