package prompts

import (
	"embed"
	"fmt"
)

//go:embed *.txt
var fs embed.FS

// Default prompt.
const Default = "You are helpful assistant."

// Get returns prompt for the specified model.
func Get(name, model string) string {
	var filename string

	// example: auth_anthropic_claude-3-7-sonnet-latest.txt
	filename = fmt.Sprintf("%s_%s.txt", name, model)
	if prompt, err := fs.ReadFile(filename); err == nil {
		return string(prompt)
	}

	// example: auth.txt
	filename = fmt.Sprintf("%s.txt", name)
	if prompt, err := fs.ReadFile(filename); err == nil {
		return string(prompt)
	}

	return Default
}
