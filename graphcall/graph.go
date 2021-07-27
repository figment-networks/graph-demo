package graphcall

type GraphQuery struct {
	Q       Part
	Queries []Query
}

type Query struct {
	Name   string
	Order  int
	Params map[string]Part
	Fields map[string]Field
}

type Part struct {
	Name   string
	Params map[string]Param
}

type Field struct {
	Name   string
	Order  int
	Params map[string]Part
	Fields map[string]Field
}

type Param struct {
	Field    string
	Type     string // TODO(lukanus): type
	Variable string
	Value    interface{}
}

// -------------------

type Subgraph struct {
	Name     string
	Entities map[string]*Entity
}

func NewSubgraph(name string) *Subgraph {
	return &Subgraph{Name: name, Entities: make(map[string]*Entity)}
}

type Entity struct {
	Name   string
	Fields map[string]Fields
}

func NewEntity(name string) *Entity {
	return &Entity{Name: name, Fields: make(map[string]Fields)}
}

type Fields struct {
	Name    string
	Type    string
	IsArray bool
	NotNull bool
}
