// Inpalce provide unified interface for updating specific values in different document
// types like json whith document structure and comments preserving.
package inplace

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/asciimoth/rewrite"
)

var (
	// ErrAbsPath is returned when a path supplied to PatchFile is absolute.
	ErrAbsPath = errors.New("path must be relative for provided root")
	// ErrDocumentLoad is returned when a document cannot be loaded.
	ErrDocumentLoad = errors.New("document loading")
	// ErrVoidKeyPath is returned when a key path is empty.
	ErrVoidKeyPath = errors.New("key path should contain at least one element")
	// ErrDocumentPatching is returned when a document cannot be patched.
	ErrDocumentPatching = errors.New("document patching")
)

// KeyPath is a slice of keys representing the path to a value.
// For example, in this JSON object the path of "value" will be []string{"A","B","C"}:
//
//	{"A":{"B":{"C":"value"}}}
type KeyPath = []string

// Document represents the source of values such as a configuration file.
// For concrete document types, see [github.com/asciimoth/inplace/json],
// [github.com/asciimoth/inplace/toml],
// [github.com/asciimoth/inplace/yaml],
// [github.com/asciimoth/inplace/regexp].
type Document interface {
	// Get returns the value of the element identified by the provided path.
	// If the element does not exist, Get should return the empty string.
	// All non-string values should be converted to their string representation.
	Get(kp KeyPath) string

	// Set changes the value of the element identified by the provided path,
	// or inserts the element if it does not exist.
	Set(kp KeyPath, value string) error

	// Save serializes the Document and returns the resulting bytes.
	Save() []byte
}

// New is the type of a [Document] constructor.
type New = func(src []byte) (Document, error)

// Patch specifies an element in a [Document] that should be changed or inserted
// and the new value for that element.
type Patch struct {
	KP    KeyPath
	Value string
}

// PatchBin constructs a new [Document] from src using con, applies all patches,
// and serializes the updated document back to bytes.
func PatchBin(con New, src []byte, patches []Patch) ([]byte, error) {
	doc, err := con(src)
	if err != nil {
		return nil, err
	}
	for _, patch := range patches {
		err := doc.Set(patch.KP, patch.Value)
		if err != nil {
			return nil, err
		}
	}
	return doc.Save(), nil
}

// PatchStr is like [PatchBin] but accepts and returns strings.
func PatchStr(con New, src string, patches []Patch) (string, error) {
	b, err := PatchBin(con, []byte(src), patches)
	if err != nil {
		return "", err
	}
	if b == nil {
		return "", err
	}
	return string(b), nil
}

// PatchFile applies patches to the file at path within the provided [os.Root].
// The path must be relative to the root (an absolute path will produce
// [ErrAbsPath]). The caller is responsible for closing the provided [os.Root]
// when done. PatchFile uses an atomic file replace when possible.
func PatchFile(con New, root *os.Root, path string, patches []Patch) error {
	if filepath.IsAbs(path) {
		return ErrAbsPath
	}

	// Read original content from the root.
	orig, err := rewrite.Read(root, path)
	if err != nil {
		return err
	}

	patched, err := PatchBin(con, orig, patches)
	if err != nil {
		return err
	}

	return rewrite.Write(root, path, patched)
}
