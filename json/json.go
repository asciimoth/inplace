// json implements [github.com/asciimoth/inplace/Document] for JSON sources.
package json

import (
	js "encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/asciimoth/inplace"
	hujs "github.com/tailscale/hujson"
)

// Compile-time assertion that *Document implements iface.Focument
var _ inplace.Document = (*Document)(nil)

// Compile-time assertion that New func implements iface.New
var _ inplace.New = New

// Compile-time assertion that NewHuJSON func implements iface.New
var _ inplace.New = NewHuJSON

// Document represents a JSON or HuJSON document.
type Document struct {
	doc    hujs.Value
	hujson bool
}

// Get returns the value of the element identified by the provided path.
// If the element does not exist, Get returns the empty string.
// Get converts all non-string values to their string representation.
func (d *Document) Get(kp inplace.KeyPath) string {
	if len(kp) < 1 {
		return ""
	}
	val := d.doc.Find(buildPointer(kp))
	if val == nil {
		return ""
	}
	return fmt.Sprint(val.Value)
}

// Set changes the value of the element identified by the provided path,
// or inserts the element if it does not exist.
func (d *Document) Set(kp inplace.KeyPath, newVal string) error {
	if len(kp) < 1 {
		return inplace.ErrVoidKeyPath
	}
	ptr := buildPointer(kp)

	val := d.doc.Find(ptr)

	if val != nil {
		patch := []map[string]any{
			{
				"op":    "replace",
				"path":  ptr,
				"value": newVal,
			},
		}
		pb, err := js.Marshal(patch)
		if err != nil {
			return errors.Join(inplace.ErrDocumentPatching, err)
		}
		if err := d.doc.Patch(pb); err != nil {
			return errors.Join(inplace.ErrDocumentPatching, err)
		}
		return nil
	}

	// pointer does not exist -> create chain via helper
	pb, err := insertValue(&d.doc, kp, newVal)
	if err != nil {
		return errors.Join(inplace.ErrDocumentPatching, err)
	}

	if err := d.doc.Patch(pb); err != nil {
		return errors.Join(inplace.ErrDocumentPatching, err)
	}
	return nil
}

// Save serializes the Document and returns the resulting bytes.
func (d *Document) Save() []byte {
	if !d.hujson {
		// Make output compilant to standadrt JSON
		d.doc.Minimize()
	}
	return d.doc.Pack()
}

// New construct [Document] from JSON.
func New(src []byte) (inplace.Document, error) {
	doc, err := hujs.Parse(src)
	if err != nil {
		return nil, errors.Join(inplace.ErrDocumentLoad, err)
	}
	return &Document{doc, false}, nil
}

// New construct [Document] from HuJSON.
func NewHuJSON(src []byte) (inplace.Document, error) {
	doc, err := hujs.Parse(src)
	if err != nil {
		return nil, errors.Join(inplace.ErrDocumentPatching, err)
	}
	return &Document{doc, true}, nil
}

// helper: build RFC6901 JSON Pointer from keyPath, with proper escaping
func buildPointer(keys inplace.KeyPath) string {
	if len(keys) == 0 {
		return "" // pointer to whole document
	}
	var sb strings.Builder
	for _, p := range keys {
		// escape ~ -> ~0 and / -> ~1
		esc := strings.ReplaceAll(p, "~", "~0")
		esc = strings.ReplaceAll(esc, "/", "~1")
		sb.WriteString("/")
		sb.WriteString(esc)
	}
	return sb.String()
}

// insertValue constructs and applies JSON Patch ops to create missing intermediate
// objects and add the final leaf. It mutates v and returns the applied patch bytes.
func insertValue(v *hujs.Value, keyPath inplace.KeyPath, newVal string) ([]byte, error) {
	if len(keyPath) == 0 {
		// nothing to insert; replace whole document
		patch := []map[string]any{
			{"op": "replace", "path": "", "value": newVal},
		}
		pb, _ := js.Marshal(patch)
		if err := v.Patch(pb); err != nil {
			return nil, fmt.Errorf("patch replace root: %w", err)
		}
		return pb, nil
	}

	// Build prefix pointers: ["/a", "/a/b", "/a/b/c"]
	prefixes := make([]string, len(keyPath))
	for i := range keyPath {
		prefixes[i] = buildPointer(keyPath[:i+1])
	}

	// find deepest existing prefix
	deepest := -1
	for i, p := range prefixes {
		if v.Find(p) != nil {
			deepest = i
		} else {
			// first missing prefix found; further prefixes are missing as well
			break
		}
	}

	var ops []map[string]any

	// For each missing prefix except the final leaf, add an empty object {}.
	// Example: to set /a/b/c where only /a exists:
	//  add /a/b = {}
	//  add /a/b/c = "newVal"
	for i := deepest + 1; i < len(prefixes)-1; i++ {
		ops = append(ops, map[string]any{
			"op":    "add",
			"path":  prefixes[i],
			"value": map[string]any{}, // create empty object
		})
	}

	// Finally add the leaf value
	leafPath := prefixes[len(prefixes)-1]
	ops = append(ops, map[string]any{
		"op":    "add",
		"path":  leafPath,
		"value": newVal,
	})

	patchBytes, err := js.Marshal(ops)
	if err != nil {
		return nil, fmt.Errorf("marshal patch: %w", err)
	}
	if err := v.Patch(patchBytes); err != nil {
		return nil, fmt.Errorf("apply patch (add chain): %w", err)
	}
	return patchBytes, nil
}
