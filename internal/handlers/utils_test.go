package handlers

import (
	"testing"
	"github.com/go-playground/validator/v10"
)

func TestFormatValidationErrors(t *testing.T) {
	v := validator.New()

	// 1. Create a "Bad" object that we know will fail
	type TestStruct struct {
		Name string `validate:"required,min=5"`
	}
	obj := TestStruct{Name: "Hi"} // Too short!

	// 2. Run validation
	err := v.Struct(obj)
	
	// 3. Run our formatting function
	formatted := formatValidationErrors(err)

	t.Logf("Formatted Map Content: %+v", formatted)

	// 4. Assert (The "Scientist" Check)
	if formatted["name"] != "too short" {
		t.Errorf("Expected 'too short', but got '%s'", formatted["name"])
	}
}