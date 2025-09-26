package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/asciimoth/inplace"
)

type opRow struct {
	Operation string   `json:"operation"` // "set","get","save"
	KeyPath   []string `json:"keypath,omitempty"`
	Value     *string  `json:"value,omitempty"`
	ValueFile *string  `json:"valuefile,omitempty"`
	Result    *string  `json:"result,omitempty"`
	// expected error string (nil => success)
	Error *string `json:"error,omitempty"`
	// alternative: expected error text read from file
	ErrorFile *string `json:"errorfile,omitempty"`
}

func (or *opRow) GetError(t *testing.T) *string {
	if or.Error != nil {
		return or.Error
	}
	if or.ErrorFile == nil {
		return nil
	}
	raw, err := os.ReadFile(*or.ErrorFile)
	if err != nil {
		t.Fatalf("failed to read cases file %v: %v", or.ErrorFile, err)
	}
	str := strings.TrimSpace(string(raw))
	return &str
}

func (or *opRow) GetValue(t *testing.T) *string {
	if or.Value != nil {
		return or.Value
	}
	raw, err := os.ReadFile(*or.ValueFile)
	if err != nil {
		t.Fatalf("failed to read cases file %v: %v", or.ValueFile, err)
	}
	str := strings.TrimSpace(string(raw))
	return &str
}

type caseRow struct {
	Name        *string `json:"name"`
	Constructor string  `json:"constructor"`
	Input       *string `json:"input,omitempty"`
	InputFile   *string `json:"inputfile,omitempty"`
	// expected constructor error (nil => success)
	Error *string `json:"error,omitempty"`
	// alternative: expected error text read from file
	ErrorFile *string `json:"errorfile,omitempty"`
	Ops       []opRow `json:"ops,omitempty"`
}

func (cr *caseRow) GetError(t *testing.T) *string {
	if cr.Error != nil {
		return cr.Error
	}
	if cr.ErrorFile == nil {
		return nil
	}
	raw, err := os.ReadFile(*cr.ErrorFile)
	if err != nil {
		t.Fatalf("failed to read cases file %v: %v", cr.ErrorFile, err)
	}
	str := strings.TrimSpace(string(raw))
	return &str
}

func (cr *caseRow) GetInput(t *testing.T) *string {
	if cr.Input != nil {
		return cr.Input
	}
	raw, err := os.ReadFile(*cr.InputFile)
	if err != nil {
		t.Fatalf("failed to read cases file %v: %v", *cr.InputFile, err)
	}
	str := strings.TrimSpace(string(raw))
	return &str
}

type testFile struct {
	Table []caseRow `json:"table"`
}

func compareError(t *testing.T, got error, want *string) {
	if want == nil {
		// expect success
		if got != nil {
			t.Fatalf("unexpected error: %v", got)
		}
		return
	}
	// expect an error string equal to want
	if got == nil {
		t.Fatalf("expected error %q but got nil", *want)
	}
	if got.Error() != *want {
		t.Fatalf("error mismatch: want %q got %q", *want, got.Error())
	}
}

func compareValue(t *testing.T, got *string, want *string) {
	if want == nil {
		// expect success
		if got != nil {
			t.Fatalf("unexpected value: %v", got)
		}
		return
	}
	if got == nil {
		t.Fatalf("expected value %q but got nil", *want)
	}
	if strings.TrimSpace(*got) != strings.TrimSpace(*want) {
		t.Fatalf("values mismatch: want %q got %q", *want, *got)
	}
}

func TestDocumentsFromJSONTable(t *testing.T, path string, constructors map[string]inplace.New) {
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cases file %q: %v", path, err)
	}

	var tf testFile
	if err := json.Unmarshal(raw, &tf); err != nil {
		t.Fatalf("invalid JSON in %s: %v", path, err)
	}
	for idx, row := range tf.Table {
		name := fmt.Sprintf("%02d-%s", idx, row.Constructor)
		if row.Name != nil {
			name = *row.Name
		}
		t.Run(name, func(t *testing.T) {
			constructor, ok := constructors[row.Constructor]
			if !ok {
				t.Fatalf("constructor %v not registered in test (update constructors map)", row.Constructor)
			}

			doc, err := constructor([]byte(*row.GetInput(t)))

			eerr := row.GetError(t)

			compareError(t, err, eerr)
			for opIdx, op := range row.Ops {
				opName := fmt.Sprintf("op%02d-%s", opIdx, op.Operation)
				t.Run(opName, func(t *testing.T) {
					eerr := op.GetError(t)
					switch op.Operation {
					case "set":
						compareError(t, doc.Set(op.KeyPath, *op.GetValue(t)), eerr)
					case "get":
						val := doc.Get(op.KeyPath)
						compareValue(t, &val, op.GetValue(t))
					case "save":
						val := string(doc.Save())
						if op.Value == nil && op.ValueFile != nil {
							_, err := os.ReadFile(*op.ValueFile)
							if err != nil {
								os.WriteFile(*op.ValueFile, []byte(val), 0o600)
								t.Fatalf("%s not presented, creating", *op.ValueFile)
							}
						}
						compareValue(t, &val, op.GetValue(t))
					default:
						t.Fatalf("unknown operation %q", op.Operation)
					}
				})
			}
		})
	}
}
