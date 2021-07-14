package graphql

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// var (
// 	partRegxp = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
// 	params    = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
// )

type GraphQuery struct {
	Q       Part
	Queries []Query
}

type Query struct {
	Name   string
	Params map[string]Part
	Fields map[string]Field
}

type Part struct {
	Name   string
	Params map[string]Param
}

type Field struct {
	Name   string
	Params map[string]Part
	Fields map[string]Field
}
type Param struct {
	Field    string
	Type     string // TODO(lukanus): type
	Variable string
	Value    interface{}
}

func NewParam(inputObjects map[string]Param, variableType ast.Type, value interface{}, field string) (p Param, err error) {
	p.Field = field

	p.Type, err = getType(variableType)
	if err != nil {
		return Param{}, err
	}

	if p.Variable, err = getVariable(inputObjects, value, p.Type); err != nil {
		return Param{}, err
	}

	if p.Value, err = getValue(inputObjects, value, field, p.Type); err != nil {
		return Param{}, err
	}

	return
}

func getType(t ast.Type) (string, error) {
	if t == nil {
		return "", nil
	}

	switch reflect.TypeOf(t) {
	case reflect.TypeOf(&ast.Named{}):
		return t.(*ast.Named).Name.Value, nil
	case reflect.TypeOf(&ast.NonNull{}):
		return getType(t.(*ast.NonNull).Type)
	case reflect.TypeOf(&ast.List{}):
		typeStr, err := getType(t.(*ast.List).Type)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("[%s]", typeStr), nil
	default:
		return "", errors.New("Unknown input type")
	}
}

func getVariable(inputParams map[string]Param, v interface{}, variableType string) (variableStr string, err error) {
	switch variableType {
	case "Int":
		variableStr = "uint64"
	case "[Int]":
		variableStr = "[]uint64"
	case "String":
		variableStr = "string"
	case "[String]":
		variableStr = "[]string"

	default:
		param, ok := inputParams[variableType]
		if !ok {
			err = fmt.Errorf("Missing input scheme for %q", variableType)
			return
		}

		variableStr = param.Variable
	}

	return
}

func getValue(inputParams map[string]Param, v interface{}, field, variableType string) (value interface{}, err error) {
	if v == nil {
		return nil, nil
	}

	switch variableType {
	case "Int":
		var float64Val float64
		if float64Val, err = float64Value(v); err != nil {
			return nil, err
		}
		return uint64Value(float64Val), nil

	case "[]Int":
		value = make([]*big.Int, len(v.([]float64)))
		for i, str := range v.([]float64) {
			float64Val, err := float64Value(str)
			if err != nil {
				return nil, err
			}
			value.([]*big.Int)[i] = uint64Value(float64Val)
		}
		return value, nil

	case "String":
		if value, err = stringValue(v); err != nil {
			return nil, err
		}
		return value, nil

	case "[]String":
		value = make([]string, len(v.([]string)))
		for i, str := range v.([]string) {
			if value.([]string)[i], err = stringValue(str); err != nil {
				return nil, err
			}
		}
		return value, nil

	default:
		param, ok := inputParams[variableType]
		if !ok {
			return nil, fmt.Errorf("Missing input scheme for %q", variableType)
		}
		param.Field = field

		for key, par := range param.Value.(map[string]Param) {
			parValue, ok := v.(map[string]interface{})[key]
			if !ok {
				return nil, fmt.Errorf("Missing input variable %q", key)
			}

			par.Value, err = getValue(inputParams, parValue, par.Field, par.Type)
			if err != nil {
				return nil, err
			}

			param.Value.(map[string]Param)[key] = par
		}

		return param.Value, nil

	}
}

func float64Value(val interface{}) (float64, error) {
	if reflect.TypeOf(val).Kind() != reflect.Float64 {
		return 0, errors.New("Value is not float64")
	}
	return val.(float64), nil
}

func stringValue(val interface{}) (string, error) {
	if reflect.TypeOf(val).Kind() != reflect.String {
		return "", errors.New("Value is not string")
	}
	return val.(string), nil
}

func (p *Param) String() (string, error) {
	if reflect.TypeOf(p.Value).Kind() != reflect.String {
		return "", errors.New("Value is not a string")
	}
	return p.Value.(string), nil
}

func ParseQuery(query string, variables map[string]interface{}) (GraphQuery, error) {
	opts := parser.ParseOptions{
		NoSource: true,
	}
	doc, err := parser.Parse(parser.ParseParams{
		Options: opts,
		Source: &source.Source{
			Body: []byte(query),
		},
	})

	if err != nil {
		return GraphQuery{}, fmt.Errorf("Error while parsing graphql query: %w", err)
	}

	q := GraphQuery{}

	inputObjects := make(map[string]Param)

	for _, definition := range doc.Definitions {
		kind := definition.GetKind()
		fmt.Println("kind: ", kind)
		switch kind {
		case "InputObjectDefinition":
			iod := definition.(*ast.InputObjectDefinition)

			objValue, err := parseInputObjectValue(inputObjects, iod.Fields)
			if err != nil {
				return GraphQuery{}, err
			}

			inputObjects[iod.Name.Value] = Param{
				Type:     "object",
				Variable: iod.Name.Value,
				Value:    objValue,
			}

		case "OperationDefinition":
			od := definition.(*ast.OperationDefinition)

			if err = q.parseOperationDefinition(od, inputObjects, variables); err != nil {
				return GraphQuery{}, err
			}

		default:
			return GraphQuery{}, errors.New("unknown graph definition")
		}
	}

	return q, nil
}

func parseInputObjectValue(inputObjects map[string]Param, fields []*ast.InputValueDefinition) (objFields map[string]Param, err error) {
	objFields = make(map[string]Param)
	for _, f := range fields {
		field := f.Name.Value

		if objFields[field], err = NewParam(inputObjects, f.Type, nil, field); err != nil {
			return nil, err
		}
	}
	return objFields, nil
}

func (q *GraphQuery) parseOperationDefinition(od *ast.OperationDefinition, inputObjects map[string]Param, variables map[string]interface{}) error {
	if od.Operation != "query" {
		return errors.New("Expected query operation")
	}

	variableDefinitions := od.VariableDefinitions
	fmt.Println("variableDefinitions ", variableDefinitions)

	// root query name
	q.Q.Name = od.Name.Value

	// root query parameters
	if err := q.queryQParams(inputObjects, variableDefinitions, variables); err != nil {
		return err
	}

	// queries
	selections := od.SelectionSet.Selections

	q.Queries = make([]Query, len(selections))
	for i, selection := range selections {
		queryField := ast.NewField(selection.(*ast.Field))
		q.Queries[i].Name = queryField.Name.Value
		arguments := queryField.Arguments

		if err := q.queryParams(arguments, i); err != nil {
			return err
		}

		fields := selection.GetSelectionSet().Selections
		q.queryFields(fields, i)
	}

	return nil
}

func (q *GraphQuery) queryQParams(inputObjects map[string]Param, variableDefinitions []*ast.VariableDefinition, variables map[string]interface{}) error {
	params := make(map[string]Param)

	for _, vd := range variableDefinitions {
		field := vd.Variable.Name.Value

		value, ok := variables[field]
		if !ok {
			return errors.New("Missing input variable")
		}

		pField, err := NewParam(inputObjects, vd.Type, value, field)
		if err != nil {
			return err
		}

		params[field] = pField
	}

	q.Q.Params = params

	return nil
}

func uint64Value(value float64) *big.Int {
	return new(big.Int).SetInt64(int64(value))
}

func (q *GraphQuery) queryParams(arguments []*ast.Argument, i int) error {
	for _, arg := range arguments {
		params := make(map[string]Part)
		argName := arg.Name.Value
		name := ast.NewName(arg.Value.GetValue().(*ast.Name))

		nameStr := name.Value

		variable := q.Q.Params[nameStr]

		params[argName] = Part{
			Name:   argName,
			Params: map[string]Param{nameStr: variable},
		}

		fmt.Println(map[string]Param{nameStr: variable})

		q.Queries[i].Params = params
	}

	return nil
}

func (q *GraphQuery) queryFields(selections []ast.Selection, i int) {
	var f Field
	fields := make(map[string]Field)

	for _, s := range selections {
		field := ast.NewField(s.(*ast.Field))
		f.Name = field.Name.Value
		fields[f.Name] = f
	}

	q.Queries[i].Fields = fields
}

func MapQueryToResponse(queries []Query) (map[string]interface{}, error) {
	rawResp := make(map[string]interface{})
	for _, query := range queries {
		fields := make(map[string]interface{})
		for _, field := range query.Fields {
			if field.Params == nil {
				return nil, errors.New("Empty response parameter value")
			}

			value, ok := field.Params[field.Name]
			if !ok {
				return nil, errors.New("Missing response parameter value")
			}

			fields[field.Name] = value
		}
		rawResp[query.Name] = fields
	}
	return rawResp, nil
}
