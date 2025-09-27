package regexp_test

import (
	"testing"

	"github.com/asciimoth/inplace"
	"github.com/asciimoth/inplace/internal/testutil"
	"github.com/asciimoth/inplace/regexp"
)

func TestDocumentsFromJSONTable(t *testing.T) {
	caseFilePath := "cases/cases.json"
	var constructors = map[string]inplace.New{
		"New": regexp.New,
	}
	testutil.TestDocumentsFromJSONTable(t, caseFilePath, constructors)
}
