package toml

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/asciimoth/inplace"
	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
	"github.com/pelletier/go-toml/v2"
)

// Compile-time assertion that *Document implements iface.Focument
var _ inplace.Document = (*Document)(nil)

// Compile-time assertion that New func implements iface.New
var _ inplace.New = New

type Document struct {
	doc tomledit.Document
}

func (d *Document) Get(kp inplace.KeyPath) string {
	if len(kp) < 1 {
		return ""
	}
	entry := d.doc.First(kp...)
	if entry == nil {
		return ""
	}
	if entry.KeyValue == nil {
		return ""
	}
	return tomlStringValue(entry.Value.String())
}

func (d *Document) Set(kp inplace.KeyPath, newVal string) error {
	if len(kp) < 1 {
		return inplace.ErrVoidKeyPath
	}
	entry := d.doc.First(kp...)
	if entry != nil {
		entry.Value = parser.MustValue(fmt.Sprintf("%q", newVal))
		return nil
	}
	return insertValue(&d.doc, kp, newVal)
}

func (d *Document) Save() []byte {
	var buf bytes.Buffer
	if err := tomledit.Format(&buf, &d.doc); err != nil {
		return []byte{}
	}
	return buf.Bytes()
}

func New(src []byte) (inplace.Document, error) {
	doc, err := tomledit.Parse(bytes.NewReader(src))
	if err != nil {
		return nil, errors.Join(inplace.ErrDocumentLoad, err)
	}
	return &Document{*doc}, nil
}

func insertValue(doc *tomledit.Document, kp inplace.KeyPath, val string) error {
	sec, kp := docFindSection(doc, kp)
	transform.InsertMapping(
		sec,
		&parser.KeyValue{
			Block: parser.Comments{},
			Name:  kp,
			Line:  0,
			Value: parser.MustValue(fmt.Sprintf("%q", val)),
		},
		true,
	)
	return nil
}

func docFindSection(doc *tomledit.Document, kp []string) (*tomledit.Section, []string) {
	if len(kp) < 2 {
		return doc.Global, kp
	}
	if doc.Sections != nil {
		for _, sec := range doc.Sections {
			if sec == nil {
				continue
			}
			if sec.Heading == nil {
				continue
			}
			if sec.Name.IsPrefixOf(kp) {
				kp = kp[len(sec.Name):]
				return sec, kp
			}
		}
	}
	return doc.Global, kp
}

func tomlStringValue(tomlStr string) string {
	doc := "v = " + tomlStr + "\n"
	var cfg struct {
		V string `toml:"v"`
	}
	if err := toml.Unmarshal([]byte(doc), &cfg); err != nil {
		return tomlStr
	}
	return cfg.V
}
