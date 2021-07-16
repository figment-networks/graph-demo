package structs

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
