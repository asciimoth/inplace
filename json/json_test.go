package json_test

import (
	"testing"

	"github.com/asciimoth/inplace"
	"github.com/asciimoth/inplace/internal/testutil"
	"github.com/asciimoth/inplace/json"
)

func TestDocumentsFromJSONTable(t *testing.T) {
	caseFilePath := "cases/cases.json"
	var constructors = map[string]inplace.New{
		"New":       json.New,
		"NewHuJSON": json.NewHuJSON,
	}

	testutil.TestDocumentsFromJSONTable(t, caseFilePath, constructors)
}
