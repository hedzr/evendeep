package tool_test

import (
	"testing"

	"github.com/hedzr/evendeep/internal/tool"
)

func TestString_Split(t *testing.T) {
	var s tool.String = "hello world"
	t.Log(s.Split(" "))
	// Output:
	// [hello world]
}

func TestTrimQuotes(t *testing.T) {
	s1 := `"Hello", World!`
	t.Logf(tool.TrimQuotes(s1), tool.StripQuotes(s1))

	s1 = `'Hello', World!`
	t.Logf(tool.TrimQuotes(s1), tool.StripQuotes(s1))

	s1 = `"Hello, World!"`
	t.Logf(tool.TrimQuotes(s1), tool.StripQuotes(s1))

	s1 = `'Hello, World!'`
	t.Logf(tool.TrimQuotes(s1), tool.StripQuotes(s1))

	s1 = `Hello, "World!"`
	t.Logf(tool.TrimQuotes(s1), tool.StripQuotes(s1))

	s1 = `Hello, 'World!'`
	t.Logf(tool.TrimQuotes(s1), tool.StripQuotes(s1))

	s1 = `Hello, World!`
	t.Logf(tool.TrimQuotes(s1), tool.StripQuotes(s1))
}
