package inplace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/asciimoth/rewrite"
)

var (
	ErrAbsPath          = fmt.Errorf("path must be relative for provided root")
	ErrDocumentLoad     = fmt.Errorf("document loading")
	ErrVoidKeyPath      = fmt.Errorf("key path should contain at least one element")
	ErrDocumentPatching = fmt.Errorf("document patching")
)

type KeyPath = []string

type Document interface {
	Get(kp KeyPath) string
	Set(kp KeyPath, value string) error
	Save() []byte
}

// Constructor for Document
type New = func(src []byte) (Document, error)

type Patch struct {
	KP    KeyPath
	Value string
}

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

func PatchStr(con New, src string, patches []Patch) (string, error) {
	b, err := PatchBin(con, []byte(src), patches)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// PatchFile applies patches to file `path` inside provided os.Root `root`.
// - `path` must be a path inside the root (not an absolute path).
// - Caller is responsible for closing the provided *os.Root when done.
//
// Uses an atomic replace: write to a temporary file inside the same root and rename it over the original.
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
