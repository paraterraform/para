package index

import (
	"fmt"
)

type Plugin struct {
	Platform string
	Name     string
	Kind     string
	Version  string
	Size     uint64
	Digest   string
	Url      string
}

func (p Plugin) Filename() string {
	return fmt.Sprintf("terraform-%s-%s_%s", p.Kind, p.Name, p.Version)
}
