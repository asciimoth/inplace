// regexp implements [github.com/asciimoth/inplace/Document] for TOML sources.
package regexp

import (
	"regexp"

	"github.com/asciimoth/inplace"
)

// Compile-time assertion that *Document implements iface.Focument
var _ inplace.Document = (*Document)(nil)

// Compile-time assertion that New func implements iface.New
var _ inplace.New = New

// Document represents arbitrary binary document that can be manipulated with
// regular expressions.
type Document struct {
	src []byte
}

// Get returns the value of the element identified by the provided path.
// Path is a series of regexps constricting the selected substring
// of the document. E.g. `[]string{"substring", "substr", "bs"}`.
// Get returns first match of kp.
// If the element does not exist, Get returns the empty string.
func (d *Document) Get(kp inplace.KeyPath) string {
	if len(kp) < 1 {
		return ""
	}
	return string(regexpRecursiveGet(d.src, kp))
}

// Set changes all occurs of kp ti newVal.
// If there is no occurs, document stays unchanged.
func (d *Document) Set(kp inplace.KeyPath, newVal string) error {
	if len(kp) < 1 {
		return inplace.ErrVoidKeyPath
	}
	d.src = regexpRecursiveReplace(d.src, kp, newVal)
	return nil
}

// Save returns raw bytes of the document.
func (d *Document) Save() []byte {
	return d.src
}

// New construct [Document] from arbitrary bytes.
func New(src []byte) (inplace.Document, error) {
	return &Document{src}, nil
}

func regexpRecursiveReplace(src []byte, patterns []string, val string) []byte {
	if len(patterns) < 1 {
		return []byte(val)
	}
	re := regexp.MustCompile(patterns[0])
	return re.ReplaceAllFunc(src, func(match []byte) []byte {
		return regexpRecursiveReplace(match, patterns[1:], val)
	})
}

func regexpRecursiveGet(src []byte, patterns []string) []byte {
	if len(patterns) < 1 {
		return []byte{}
	}
	// Precompile regexps. If any fails to compile, treat that as "no match".
	regs := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		r, err := regexp.Compile(p)
		if err != nil {
			return []byte{}
		}
		regs[i] = r
	}

	var search func(cur []byte, idx int) []byte
	search = func(cur []byte, idx int) []byte {
		// If we've consumed all patterns, current substring is the "full match".
		if idx >= len(regs) {
			return cur
		}

		r := regs[idx]
		// Find all non-overlapping matches of this pattern in cur (in order).
		matches := r.FindAllIndex(cur, -1)
		if matches == nil {
			return []byte{}
		}

		for _, mm := range matches {
			start, end := mm[0], mm[1]
			sub := cur[start:end]
			if res := search(sub, idx+1); res != nil {
				return res // return first successful full match
			}
			// otherwise continue to next match of this level
		}
		return []byte{}
	}

	return search(src, 0)
}
