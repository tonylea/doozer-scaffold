package prompt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tonylea/doozer-scaffold/internal/prompt"
)

func TestSanitiseForIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-project", "my_project"},
		{"MyApp", "myapp"},
		{"123-bad-start", "bad_start"},
		{"---", "app"},
		{"hello_world", "hello_world"},
		{"UPPER", "upper"},
		{"a", "a"},
		{"", "app"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, prompt.SanitiseForIdentifier(tc.input))
		})
	}
}
