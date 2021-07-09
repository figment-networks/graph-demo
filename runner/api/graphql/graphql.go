package graphql

import (
	"log"
	"regexp"

	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

var (
	partRegxp = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
	params    = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
)

type Query struct {
	Q      Part
	Params map[string]Part
	Fields map[string]Field
}

type Part struct {
	Name   string
	Params map[string]Param
}

type Param struct {
	Field    string
	Type     string // TODO(lukanus): type
	Variable string
	Value    interface{}
}

type Field struct {
	Name   string
	Params map[string]Part

	Fields map[string]Field
}

func ParseQuery(query string, variables map[string]interface{}) (q Query, err error) {
	opts := parser.ParseOptions{
		NoSource: true,
	}
	a, err := parser.Parse(parser.ParseParams{
		Options: opts,
		Source: &source.Source{
			Body: []byte(query),
		},
	})
	log.Println("a", a, err)
	return q, nil

}

/*

func ParseQuery(query string, variables map[string]interface{}) (q Query, err error) {
	q = Query{}
	pos := strings.Index(query, "{")
	if pos < 0 {
		return q, errors.New("invalid graphql")
	}

	initialPortion := query[0 : pos+1]
	if strings.Contains(initialPortion, "query") {
		parts := partRegxp.FindAllStringSubmatch(initialPortion+"\n", -1)
		log.Println("parts", parts)
		if len(parts) == 1 {
			parsedP, err := parseParams(parts[0][2])
			if err != nil {
				return q, err
			}
			q.Q = Part{
				Name:   parts[0][1],
				Params: parsedP,
			}
		}
	}

	parts := partRegxp.FindAllStringSubmatch(query[pos+1:], -1)
	for _, p := range parts {

		parsedP, err := parseParams(parts[0][2])
		if err != nil {
			return q, err
		}

		//	q.Params[] = Part{
	//		Name:   parts[0][1],
	//		Params: parsedP,
	//	}
	}

	return q, nil
}

func parseParams(paramS string) (map[string]Param, error) {
	if paramS[0] == '(' {
		paramS = paramS[1:]
	}

	if paramS[len(paramS)-1] == ')' {
		paramS = paramS[:len(paramS)-2]
	}

	paramsA := strings.Split(paramS, ",")
	params := make(map[string]Param)

	for _, p := range paramsA {
		paramPart := strings.Split(p, ":")
		if len(paramPart) < 2 {
			return nil, errors.New("params in wrong format")
		}
		val := strings.Trim(paramPart[1], " ")

		param := Param{
			Field: strings.Trim(paramPart[0], " "),
		}
		if val[0] == '$' {
			param.Variable = val[1:]
		} else {
			if val[0] == '"' {
				param.Type = "string"
				param.Value = val[1 : len(val)-2]
			} else {
				param.Value = val
			}
		}

		params[param.Field] = param
	}

	return params, nil
}

func parseQueryParams(paramS string) (map[string]Param, error) {
	if paramS[0] == '(' {
		paramS = paramS[1:]
	}

	if paramS[len(paramS)-1] == ')' {
		paramS = paramS[:len(paramS)-2]
	}

	paramsA := strings.Split(paramS, ",")
	params := make(map[string]Param)

	for _, p := range paramsA {
		paramPart := strings.Split(p, ":")
		if len(paramPart) < 2 {
			return nil, errors.New("params in wrong format")
		}
		val := strings.Trim(paramPart[1], " ")

		param := Param{
			Field: strings.Trim(paramPart[0], " "),
		}
		if val[0] == '$' {
			param.Variable = val[1:]
		} else {
			if val[0] == '"' {
				param.Type = "string"
				param.Value = val[1 : len(val)-2]
			} else {
				param.Value = val
			}
		}

		params[param.Field] = param
	}

	return params, nil
}

func ParseInnerQuery(query []string, variables map[string]interface{}) (q Query, err error) {

	for _, v := range v {

	}
}
*/
