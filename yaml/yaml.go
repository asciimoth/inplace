package yaml

import (
	"errors"
	"fmt"

	"github.com/asciimoth/inplace"
	yaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
)

// Compile-time assertion that *Document implements iface.Focument
var _ inplace.Document = (*Document)(nil)

// Compile-time assertion that New func implements iface.New
var _ inplace.New = New

type Document struct {
	ast ast.File
}

func (d *Document) Get(kp inplace.KeyPath) string {
	if len(kp) < 1 {
		return ""
	}
	path := keysToYamlPath(kp)

	node, err := path.FilterFile(&d.ast)
	if err != nil {
		return ""
	}
	return node.GetToken().Value
}

func (d *Document) Set(kp inplace.KeyPath, newVal string) error {
	if len(kp) < 1 {
		return inplace.ErrVoidKeyPath
	}
	path := keysToYamlPath(kp)

	// Check if value with this key path exist
	_, err := path.FilterFile(&d.ast)

	// If there is no such key path, instert
	if errors.Is(err, yaml.ErrNotFoundNode) {
		err = insertNestedValueAST(&d.ast, kp, newVal)
		if err != nil {
			return errors.Join(inplace.ErrDocumentPatching, err)
		}
		return nil
	}

	if err != nil {
		return errors.Join(inplace.ErrDocumentPatching, err)
	}

	// Repalce value with new one
	err = path.ReplaceWithNode(
		&d.ast,
		ast.String(token.String(newVal, newVal, &token.Position{})),
	)
	if err != nil {
		return errors.Join(inplace.ErrDocumentPatching, err)
	}
	return nil
}

func (d *Document) Save() []byte {
	return []byte(d.ast.String())
}

func New(src []byte) (inplace.Document, error) {
	ast, err := parser.ParseBytes([]byte(src), parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return &Document{ast: *ast}, nil
}

func keysToYamlPath(keys []string) *yaml.Path {
	var pb yaml.PathBuilder
	pb = *pb.Root() // start from root ($)
	for _, k := range keys {
		pb = *pb.Child(k)
	}
	return pb.Build()
}

// Very dirty code
// TODO: rewrite
func insertNestedValueAST(fileAST *ast.File, keyPath []string, newVal string) error {
	// Find deepest existing prefix
	deepestIndex := -1
	var deepestPath *yaml.Path
	for i := range keyPath {
		var pb yaml.PathBuilder
		pb = *pb.Root()
		for j := 0; j <= i; j++ {
			pb = *pb.Child(keyPath[j])
		}
		p := pb.Build()
		n, _ := p.FilterFile(fileAST)
		if n != nil {
			deepestIndex = i
			deepestPath = p
		} else {
			// first missing prefix reached -> stop scanning deeper
			break
		}
	}

	tailIdx := deepestIndex + 1
	if tailIdx >= len(keyPath) {
		return fmt.Errorf("no tail to merge (unexpected)")
	}

	// Build nested Go value representing the missing tail with final value.
	var nested any = newVal
	for i := len(keyPath) - 1; i >= tailIdx; i-- {
		nested = map[string]any{keyPath[i]: nested}
	}

	// Convert that Go value into an AST node.
	nestedNode, err := yaml.ValueToNode(nested)
	if err != nil {
		return fmt.Errorf("value->node for nested tail: %w", err)
	}

	// Merge into root or deepest existing prefix
	if deepestIndex == -1 {
		rootPath := (&yaml.PathBuilder{}).Root().Build()
		if err := rootPath.MergeFromNode(fileAST, nestedNode); err != nil {
			return fmt.Errorf("merge into root: %w", err)
		}
	} else {
		if err := deepestPath.MergeFromNode(fileAST, nestedNode); err != nil {
			return fmt.Errorf("merge into prefix %v: %w", keyPath[:deepestIndex+1], err)
		}
	}
	return nil
}
