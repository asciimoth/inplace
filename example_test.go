package inplace_test

import (
	"fmt"

	"github.com/asciimoth/inplace"
	"github.com/asciimoth/inplace/toml"
)

func Example_cargo_toml() {
	patched, _ := inplace.PatchStr(
		toml.New,
		`
			[package]
			name = "name"
			version = "1.2.3"
			edition = "2018"
			authors = ["JohnDoe"]
			description = "DESCRIPTION"
			repository = "https://example.com/jdoe/name"
			readme = "README.md"
			license = "WTFPL"

			[dependencies]
			name1 = "0.22"
			name2 = "1"
			name3 = {version = "4.5.6", features = ["feature"]}
		`,
		[]inplace.Patch{
			{
				KP:    []string{"package", "name"},
				Value: "NAME",
			},
			{
				KP:    []string{"package", "version"},
				Value: "6.6.6",
			},
			{
				KP:    []string{"dependencies", "name3", "version"},
				Value: "6.6.6",
			},
		},
	)
	fmt.Println(patched)

	// Output:
	// [package]
	// name = "NAME"
	// version = "6.6.6"
	// edition = "2018"
	// authors = ["JohnDoe"]
	// description = "DESCRIPTION"
	// repository = "https://example.com/jdoe/name"
	// readme = "README.md"
	// license = "WTFPL"
	//
	// [dependencies]
	// name1 = "0.22"
	// name2 = "1"
	// name3 = {version = "6.6.6", features = ["feature"]}
}
