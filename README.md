# inplace
Go library for document's content modification

```go
import (
    "os"

	"github.com/asciimoth/inplace"
	"github.com/asciimoth/inplace/toml"
)

func main() {
    root, _ := os.OpenRoot(".")
    _ = inplace.PatchFile(
        toml.New,
        root,
        "Cargo.toml",
		[]inplace.Patch{
			{
				KP:    []string{"package", "name"},
				Value: "NEWNAME",
			},
			{
				KP:    []string{"package", "version"},
				Value: "3.1.4",
			},
		},
    )
}
```

Note that this library created for specific cases with simple scalar values manipulation.  
If you need more complex ast manipulation check this libs:
- [tailscale/hujson](github.com/tailscale/hujson)
- [creachadair/tomledit](github.com/creachadair/tomledit)
- [goccy/go-yaml](github.com/goccy/go-yaml)

If you need to just marshal/unmarshal formats to go data structures check this:
- [stdlib encoding/json](https://pkg.go.dev/encoding/json)
- [pelletier/go-toml](github.com/pelletier/go-toml)
- [goccy/go-yaml](github.com/goccy/go-yaml)

