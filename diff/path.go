package diff

import "strings"

type Path struct {
	parts []PathPart
}

func (dp Path) appendAndNew(parts ...PathPart) Path {
	return Path{parts: append(dp.parts, parts...)}
}

func (dp Path) String() string {
	var sb strings.Builder
	for _, p := range dp.parts {
		if sb.Len() > 0 {
			_, _ = sb.WriteRune('.')
		}
		_, _ = sb.WriteString(p.String())
	}
	return sb.String()
}

type PathPart interface {
	String() string
}
