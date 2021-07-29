package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/figment-networks/graph-demo/graphcall"
	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/google/uuid"
)

func mapBlocksToResponse(queries []graphcall.Query, qResp structs.QueriesResp) ([]byte, error) {
	var resp mapSlice
	var err error

	resp = make([]mapItem, len(queries))
	for _, query := range queries {

		blocks, ok := qResp[query.Name]
		if !ok {
			return nil, errors.New("response is empty")
		}

		bLen := len(blocks)
		var response interface{}
		responses := make([]interface{}, bLen)
		i := 0
		for _, bAndTxs := range blocks {
			if response, err = fieldsResp(&query, bAndTxs); err != nil {
				return nil, err
			}

			responses[i] = response
			i++
		}

		if bLen > 1 {
			response = responses
		}

		resp[query.Order] = mapItem{
			Key:   query.Name,
			Value: response,
		}
	}

	return resp.marshalJSON()
}

func fieldsResp(q *graphcall.Query, blockAndTx structs.BlockAndTx) (resp interface{}, err error) {
	qStr := strings.ToLower(q.Name)
	parseBlock := strings.Contains(qStr, "block")
	parseTransaction := strings.Contains(qStr, "transaction")

	if parseBlock && !parseTransaction {
		resp = mapStructToFields(q.Fields, blockAndTx.Block)
	} else if !parseBlock && parseTransaction {
		resp = mapSliceToFields(q.Fields, blockAndTx.Transactions)
	} else {
		resp = mapStructToFields(q.Fields, blockAndTx)

	}

	return resp, err
}

func mapStructToFields(fields map[string]graphcall.Field, s interface{}) mapSlice {
	var value interface{}
	v := reflect.Indirect(reflect.ValueOf(s))
	respMap := make(map[int]mapItem)
	maxOrder := 0

	for i := 0; i < v.NumField(); i++ {
		fieldName := strings.ToLower(v.Type().Field(i).Name)

		if nameIsStrict(fieldName) {
			continue
		}

		field, ok := fields[fieldName]
		if !ok {
			// omit fields that are not defined in the graph query
			continue
		}

		filedValue := v.Field(i).Interface()
		fieldType := reflect.TypeOf(filedValue)
		fieldKind := fieldType.Kind()

		switch fieldType {
		case reflect.TypeOf(time.Time{}), reflect.TypeOf([][]uint8{}):
			value = formatValue(filedValue)
		default:
			switch fieldKind {
			case reflect.Ptr:
				if filedValue = v.Field(i).Elem().Interface(); filedValue != nil {
					value = mapStructToFields(field.Fields, filedValue)
				}
			case reflect.Slice:
				value = mapSliceToFields(field.Fields, filedValue)
			case reflect.Struct:
				value = mapStructToFields(field.Fields, filedValue)
			default:
				value = formatValue(filedValue)
			}
		}

		order := field.Order

		respMap[order] = mapItem{
			Key:   field.Name,
			Value: value,
		}

		if maxOrder < order {
			maxOrder = order
		}

	}

	respLen := len(respMap)
	ms := make([]mapItem, respLen)

	i := 0
	for order := 0; order <= maxOrder; order++ {
		resp, ok := respMap[order]
		if ok && resp.Key != nil {
			ms[i] = resp
			i++
		}
	}

	if i == 0 {
		return nil
	}

	return ms
}

func nameIsStrict(name string) bool {
	return name == "id"
}

func formatValue(v interface{}) (val interface{}) {
	switch reflect.TypeOf(v) {
	case reflect.TypeOf(uuid.UUID{}):
		val = v.(uuid.UUID).String()
	case reflect.TypeOf(time.Time{}):
		val = v.(time.Time).Unix()
	case reflect.TypeOf([][]uint8{}):
		value := v.([][]uint8)
		sliceVal := make([]string, len(value))
		for i, byteSlice := range value {
			sliceVal[i] = fmt.Sprintf("%x", byteSlice)
		}
		val = sliceVal
	default:
		val = v
	}

	return
}

func mapSliceToFields(fields map[string]graphcall.Field, s interface{}) []mapSlice {
	v := reflect.Indirect(reflect.ValueOf(s))
	len := v.Len()
	sliceResp := make([]mapSlice, len)

	for i := 0; i < len; i++ {
		sliceResp[i] = mapStructToFields(fields, v.Index(i).Interface())
	}

	return sliceResp
}

type mapItem struct {
	Key, Value interface{}
}

type mapSlice []mapItem

func (ms mapSlice) marshalJSON() ([]byte, error) {
	var b []byte
	var err error
	buf := &bytes.Buffer{}

	buf.Write([]byte{'{'})

	for i, mi := range ms {

		switch reflect.ValueOf(mi.Value).Type().String() {
		case reflect.ValueOf([]mapSlice{}).Type().String():
			sliceValue := mi.Value.([]mapSlice)
			sliceLen := len(sliceValue)
			sliceBytes := []byte{'['}
			for i, ms := range sliceValue {
				valueBytes, err := ms.marshalJSON()
				if err != nil {
					return nil, err
				}
				sliceBytes = append(sliceBytes, valueBytes...)
				if i < sliceLen-1 {
					sliceBytes = append(sliceBytes, ',')
				}
			}
			b = append(sliceBytes, ']')

		case reflect.ValueOf(mapSlice{}).Type().String():
			b, err = mi.Value.(mapSlice).marshalJSON()
		default:
			b, err = json.Marshal(&mi.Value)
		}

		if err != nil {
			return nil, err
		}

		buf.WriteString(fmt.Sprintf("%q:", fmt.Sprintf("%v", mi.Key)))
		buf.Write(b)

		if i < len(ms)-1 {
			buf.Write([]byte{','})
		}

	}

	buf.Write([]byte{'}'})

	return buf.Bytes(), nil
}
