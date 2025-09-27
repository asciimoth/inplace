package yaml_test

import (
	"testing"

	"github.com/asciimoth/inplace"
	"github.com/asciimoth/inplace/internal/testutil"
	"github.com/asciimoth/inplace/yaml"
)

func TestDocumentsFromJSONable(t *testing.T) {
	caseFilePath := "cases/cases.json"
	var constructors = map[string]inplace.New{
		"New": yaml.New,
	}
	testutil.TestDocumentsFromJSONTable(t, caseFilePath, constructors)
}
