package graphcall

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// var (
// 	partRegxp = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
// 	params    = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
// )

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

			if err = parseOperationDefinition(&q, od, inputObjects, variables); err != nil {
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

		if objFields[field], err = newParam(inputObjects, f.Type, nil, field); err != nil {
			return nil, err
		}
	}
	return objFields, nil
}

func newParam(inputObjects map[string]Param, variableType ast.Type, value interface{}, field string) (p Param, err error) {
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
		return uint64(float64Val), nil

	case "[]Int":
		value = make([]uint64, len(v.([]float64)))
		for i, str := range v.([]float64) {
			float64Val, err := float64Value(str)
			if err != nil {
				return nil, err
			}
			value.([]uint64)[i] = uint64(float64Val)
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

func parseOperationDefinition(q *GraphQuery, od *ast.OperationDefinition, inputObjects map[string]Param, variables map[string]interface{}) (err error) {
	if od.Operation != "query" {
		return errors.New("Expected query operation")
	}

	variableDefinitions := od.VariableDefinitions

	// root query name
	q.Q.Name = od.Name.Value

	// root query parameters
	if err = queryQParams(q, inputObjects, variableDefinitions, variables); err != nil {
		return err
	}

	// queries
	selections := od.SelectionSet.Selections

	q.Queries = make([]Query, len(selections))
	for i, selection := range selections {
		queryField := ast.NewField(selection.(*ast.Field))
		q.Queries[i].Name = queryField.Name.Value
		q.Queries[i].Order = i

		arguments := queryField.Arguments

		if q.Queries[i].Params, err = queryParams(q.Q.Params, arguments); err != nil {
			return err
		}

		fields := selection.GetSelectionSet().Selections
		q.Queries[i].Fields = queryFields(fields)
	}

	return nil
}

func queryQParams(q *GraphQuery, inputObjects map[string]Param, variableDefinitions []*ast.VariableDefinition, variables map[string]interface{}) (err error) {
	params := make(map[string]Param)

	for _, vd := range variableDefinitions {
		field := vd.Variable.Name.Value
		value, ok := variables[field]
		if !ok {
			value, err = getQueryValue(vd.DefaultValue)
			if err != nil {
				return err
			}
		}

		pField, err := newParam(inputObjects, vd.Type, value, field)
		if err != nil {
			return err
		}

		params[field] = pField
	}

	q.Q.Params = params

	return nil
}

func queryParams(inputParams map[string]Param, arguments []*ast.Argument) (params map[string]Part, err error) {
	params = make(map[string]Part)
	for _, arg := range arguments {
		var value interface{}
		argName := arg.Name.Value
		argValue := arg.Value.GetValue()

		nameStr := argName

		switch arg.Value.GetKind() {
		case "IntValue":
			intValue, err := strconv.Atoi(argValue.(string))
			if err != nil {
				return nil, err
			}
			value = float64(intValue)
		case "Variable", "Name":
			nameStr = argValue.(*ast.Name).Value
		default:
			value = argValue
		}

		variable := inputParams[nameStr]
		if value != nil {
			variable.Value = value
		}

		params[argName] = Part{
			Name:   argName,
			Params: map[string]Param{nameStr: variable},
		}
	}

	return params, nil
}

func getQueryValue(value ast.Value) (interface{}, error) {
	val := value.GetValue()

	switch value.GetKind() {
	case "IntValue":
		intValue, err := strconv.Atoi(val.(string))
		if err != nil {
			return nil, err
		}
		return float64(intValue), nil
	default:
		return val, nil
	}
}

func queryFields(selections []ast.Selection) map[string]Field {
	fields := make(map[string]Field)

	for i, s := range selections {

		switch reflect.TypeOf(s) {
		case reflect.TypeOf(&ast.Field{}):
			var f Field
			field := ast.NewField(s.(*ast.Field))
			f.Name = field.Name.Value
			f.Order = i

			if field.SelectionSet != nil {
				f.Fields = queryFields(field.SelectionSet.Selections)
			}
			fields[f.Name] = f
		case reflect.TypeOf(&ast.FragmentSpread{}):
			fs := ast.NewFragmentSpread(s.(*ast.FragmentSpread))
			fmt.Println(fs)
		}

	}

	return fields
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

// func uint64Value(value float64) *big.Int {
// 	return new(big.Int).SetInt64(int64(value))
// }
