package schema

import (
	"io/ioutil"
	"regexp"

	"github.com/figment-networks/graph-demo/graphcall"
)

var (
	entityRegxp = regexp.MustCompile("type\\s+([^\\s]+)\\s+\\@entity\\s+{([^\\}]+)}")
	kvRegxp     = regexp.MustCompile("\\s+([a-zA-Z0-9]+):\\s+([\\[\\]a-zA-Z0-9]+)!?")
)

type Schemas struct {
	Subgraphs map[string]*graphcall.Subgraph
}

func NewSchemas() *Schemas {
	return &Schemas{make(map[string]*graphcall.Subgraph)}
}

func (s *Schemas) LoadFromFile(name, path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	sg, err := graphcall.ParseSchema(name, f)
	if err != nil {
		return err
	}
	s.Subgraphs[name] = sg

	return nil

}
