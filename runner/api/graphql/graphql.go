package graphql

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/figment-networks/graph-demo/runner/api/structs"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// var (
// 	structs.partRegxp = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
// 	params    = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
// )

// func (p *structs.Param) String() (string, error) {
// 	if reflect.TypeOf(p.Value).Kind() != reflect.String {
// 		return "", errors.New("Value is not a string")
// 	}
// 	return p.Value.(string), nil
// }

func ParseQuery(query string, variables map[string]interface{}) (structs.GraphQuery, error) {
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
		return structs.GraphQuery{}, fmt.Errorf("Error while parsing graphql query: %w", err)
	}

	q := structs.GraphQuery{}

	inputObjects := make(map[string]structs.Param)

	for _, definition := range doc.Definitions {
		kind := definition.GetKind()
		fmt.Println("kind: ", kind)
		switch kind {
		case "InputObjectDefinition":
			iod := definition.(*ast.InputObjectDefinition)

			objValue, err := parseInputObjectValue(inputObjects, iod.Fields)
			if err != nil {
				return structs.GraphQuery{}, err
			}

			inputObjects[iod.Name.Value] = structs.Param{
				Type:     "object",
				Variable: iod.Name.Value,
				Value:    objValue,
			}

		case "OperationDefinition":
			od := definition.(*ast.OperationDefinition)

			if err = parseOperationDefinition(&q, od, inputObjects, variables); err != nil {
				return structs.GraphQuery{}, err
			}

		default:
			return structs.GraphQuery{}, errors.New("unknown graph definition")
		}
	}

	return q, nil
}

func parseInputObjectValue(inputObjects map[string]structs.Param, fields []*ast.InputValueDefinition) (objFields map[string]structs.Param, err error) {
	objFields = make(map[string]structs.Param)
	for _, f := range fields {
		field := f.Name.Value

		if objFields[field], err = NewParam(inputObjects, f.Type, nil, field); err != nil {
			return nil, err
		}
	}
	return objFields, nil
}

func MapBlocksToResponse(queries []structs.Query, blocksResp structs.QueriesResp) (resp map[string]interface{}, err error) {
	resp = make(map[string]interface{})

	for _, query := range queries {
		blocks, ok := blocksResp[query.Name]
		if !ok {
			return nil, errors.New("Response is empty")
		}

		bLen := len(blocks)
		row := map[string]interface{}{}
		rows := make([]interface{}, bLen)
		i := 0
		for _, bAndTxs := range blocks {
			if row, err = fieldsResp(&query, bAndTxs); err != nil {
				return nil, err
			}

			rows[i] = row
			i++
		}

		if bLen == 1 {
			resp[query.Name] = row
		} else {
			resp[query.Name] = rows
		}
	}

	return resp, nil
}

func fieldsResp(q *structs.Query, blockAndTx structs.BlockAndTx) (resp map[string]interface{}, err error) {
	resp = make(map[string]interface{})

	qStr := strings.ToLower(q.Name)
	parseBlock := strings.Contains(qStr, "block")
	parseTransaction := strings.Contains(qStr, "transaction")

	if parseBlock {
		mapStructToFields(resp, q.Fields, blockAndTx.Block)
		// blockFieldsMap["block"] = mapStructToFields(q.Fields, blockAndTx.Block)
	}

	if parseTransaction {
		// blockFieldsMap["txs"] = mapStructToFields(q.Fields, blockAndTx.Txs)
	}

	return resp, err
}

// func fields(q *structs.Query, v interface{}) map[string]interface{} {
// 	resp := make(map[string]interface{})
// 	name := reflect.ValueOf(v).Type().Name()

// 	switch reflect.TypeOf(v).Kind() {
// 	case reflect.Slice:
// 		resp[name] = mapSliceToFields(q, v)
// 	case reflect.Struct:
// 		resp[name] = mapStructToFields(q, v)
// 	default:
// 		resp[name] = v
// 	}

// 	return resp
// }

func mapStructToFields(resp map[string]interface{}, fields map[string]structs.Field, s interface{}) {

	v := reflect.Indirect(reflect.ValueOf(s))

	for i := 0; i < v.NumField(); i++ {
		fieldName := strings.ToLower(v.Type().Field(i).Name)

		field, ok := fields[fieldName]
		if !ok {
			// omit fields that are not defined in the graph query
			continue
		}

		// for _, f := range field.Fields {
		// 	f.Name
		// }
		resp[field.Name] = formatValue(v.Field(i).Interface())

		// resp[f.Name] =

		// field := v.Field(i)

		// fmt.Println("kind", reflect.TypeOf(field).Kind())

		// switch reflect.TypeOf(field).Kind() {
		// case reflect.Slice:
		// 	resp[fieldName] = mapSliceToFields(fields, field.Interface())
		// case reflect.Struct:
		// 	resp[fieldName] = mapStructToFields(fields, field.Interface())
		// default:
		// 	resp[fieldName] = field.Interface()
		// }
	}
}

func formatValue(v interface{}) (val interface{}) {
	fmt.Println(reflect.TypeOf(v).Kind())
	fmt.Println(reflect.TypeOf(v))

	switch reflect.TypeOf(v) {
	case reflect.TypeOf(uuid.UUID{}):
		val = v.(uuid.UUID).String()
	case reflect.TypeOf(time.Time{}):
		val = v.(time.Time).Unix()
	default:
		val = v
	}

	return
}

func mapSliceToFields(fields map[string]structs.Field, s interface{}) []interface{} {
	v := reflect.Indirect(reflect.ValueOf(s))
	len := v.Len()
	sliceResp := make([]interface{}, len)

	for i := 0; i < len; i++ {
		fieldName := v.Type().Field(i).Name

		fmt.Println(fieldName)

		// field, ok := fields[fieldName]
		// if !ok {
		// 	// omit fields that are not defined in the graph query
		// 	continue
		// }

		// // for _, f := range field.Fields {
		// // 	f.Name
		// // }
		// resp[field.Name] = v.Field(i).Interface()

		// fmt.Println(v.Slice(i, len-i).Interface())
		// // v.Slice(i, len-i)

		// fieldName := v.Type().Field(i).Name

		// field := v.Field(i)

		// switch reflect.TypeOf(field).Kind() {
		// case reflect.Slice:

		// case reflect.Struct:
		// 	structResp := make(map[string]interface{})
		// 	mapStructToFields(field.Interface(), structResp)
		// 	resp[fieldName] = structResp
		// default:
		// 	resp[fieldName] = field.Interface()
		// }
	}

	return sliceResp
}

func NewParam(inputObjects map[string]structs.Param, variableType ast.Type, value interface{}, field string) (p structs.Param, err error) {
	p.Field = field

	p.Type, err = getType(variableType)
	if err != nil {
		return structs.Param{}, err
	}

	if p.Variable, err = getVariable(inputObjects, value, p.Type); err != nil {
		return structs.Param{}, err
	}

	if p.Value, err = getValue(inputObjects, value, field, p.Type); err != nil {
		return structs.Param{}, err
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

func getVariable(inputParams map[string]structs.Param, v interface{}, variableType string) (variableStr string, err error) {
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

func getValue(inputParams map[string]structs.Param, v interface{}, field, variableType string) (value interface{}, err error) {
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

		for key, par := range param.Value.(map[string]structs.Param) {
			parValue, ok := v.(map[string]interface{})[key]
			if !ok {
				return nil, fmt.Errorf("Missing input variable %q", key)
			}

			par.Value, err = getValue(inputParams, parValue, par.Field, par.Type)
			if err != nil {
				return nil, err
			}

			param.Value.(map[string]structs.Param)[key] = par
		}

		return param.Value, nil

	}
}

func parseOperationDefinition(q *structs.GraphQuery, od *ast.OperationDefinition, inputObjects map[string]structs.Param, variables map[string]interface{}) error {
	if od.Operation != "query" {
		return errors.New("Expected query operation")
	}

	variableDefinitions := od.VariableDefinitions
	fmt.Println("variableDefinitions ", variableDefinitions)

	// root query name
	q.Q.Name = od.Name.Value

	// root query parameters
	if err := queryQParams(q, inputObjects, variableDefinitions, variables); err != nil {
		return err
	}

	// queries
	selections := od.SelectionSet.Selections

	q.Queries = make([]structs.Query, len(selections))
	for i, selection := range selections {
		queryField := ast.NewField(selection.(*ast.Field))
		q.Queries[i].Name = queryField.Name.Value
		arguments := queryField.Arguments

		if err := queryParams(q, arguments, i); err != nil {
			return err
		}

		fields := selection.GetSelectionSet().Selections
		queryFields(q, fields, i)
	}

	return nil
}

func queryQParams(q *structs.GraphQuery, inputObjects map[string]structs.Param, variableDefinitions []*ast.VariableDefinition, variables map[string]interface{}) error {
	params := make(map[string]structs.Param)

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

func queryParams(q *structs.GraphQuery, arguments []*ast.Argument, i int) error {
	for _, arg := range arguments {
		params := make(map[string]structs.Part)
		argName := arg.Name.Value
		name := ast.NewName(arg.Value.GetValue().(*ast.Name))

		nameStr := name.Value

		variable := q.Q.Params[nameStr]

		params[argName] = structs.Part{
			Name:   argName,
			Params: map[string]structs.Param{nameStr: variable},
		}

		fmt.Println(map[string]structs.Param{nameStr: variable})

		q.Queries[i].Params = params
	}

	return nil
}

func queryFields(q *structs.GraphQuery, selections []ast.Selection, i int) {
	var f structs.Field
	fields := make(map[string]structs.Field)

	for _, s := range selections {
		field := ast.NewField(s.(*ast.Field))
		f.Name = field.Name.Value
		fields[f.Name] = f
	}

	q.Queries[i].Fields = fields
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

func uint64Value(value float64) *big.Int {
	return new(big.Int).SetInt64(int64(value))
}
