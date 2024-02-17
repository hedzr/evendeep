package tool

import "strings"

type String string

func (s String) Split(sep string) []string {
	return strings.Split(string(s), sep)
}

// StripQuotes strips first and last quote char (double quote or single quote).
func StripQuotes(s string) string { return trimQuotes(s) }

// TrimQuotes strips first and last quote char (double quote or single quote).
func TrimQuotes(s string) string { return trimQuotes(s) }

// func trimQuotes(s string) string {
// 	if len(s) >= 2 {
// 		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
// 			return s[1 : len(s)-1]
// 		}
// 	}
// 	return s
// }

func trimQuotes(s string) string {
	switch {
	case s[0] == '\'':
		if s[len(s)-1] == '\'' {
			return s[1 : len(s)-1]
		}
		return s[1:]
	case s[0] == '"':
		if s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
		return s[1:]
	case s[len(s)-1] == '\'':
		return s[0 : len(s)-1]
	case s[len(s)-1] == '"':
		return s[0 : len(s)-1]
	}
	return s
}
