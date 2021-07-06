package schema

import (
	"io/ioutil"
	"regexp"
)

var (
	entityRegxp = regexp.MustCompile("type\\s+([^\\s]+)\\s+\\@entity\\s+{([^\\}]+)}")
	kvRegxp     = regexp.MustCompile("\\s+([a-zA-Z0-9]+):\\s+([\\[\\]a-zA-Z0-9]+)!?")
)

type Subgraph struct {
	Name     string
	Entities map[string]*Entity
}

func NewSubgraph(name string) *Subgraph {
	return &Subgraph{Name: name, Entities: make(map[string]*Entity)}
}

type Entity struct {
	Name   string
	Params map[string]Param
}

type Param struct {
	Type    string
	IsArray bool
}

func NewEntity(name string) *Entity {
	return &Entity{Name: name, Params: make(map[string]Param)}
}

type Schemas struct {
	Subgraphs map[string]*Subgraph
}

func NewSchemas() *Schemas {
	return &Schemas{make(map[string]*Subgraph)}
}

func (s *Schemas) LoadFromFile(name, path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	sg := NewSubgraph(name)

	declarations := entityRegxp.FindAllSubmatch(f, -1)

	for _, v := range declarations {
		if len(v) == 3 {
			ent := NewEntity(string(v[1]))
			params := kvRegxp.FindAllSubmatch(v[2], -1)
			for _, p := range params {
				if len(p) == 3 {
					typeS := string(p[2])
					if typeS[0] == '[' { // set
						ent.Params[string(p[1])] = Param{Type: string(typeS[1 : len(typeS)-1])}
					} else {
						ent.Params[string(p[1])] = Param{Type: string(typeS)}
					}

				}
			}
			sg.Entities[ent.Name] = ent
		}
	}

	s.Subgraphs[name] = sg
	return nil

}
