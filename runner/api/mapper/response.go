package mapper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/figment-networks/graph-demo/runner/api/structs"
	"github.com/google/uuid"
)

func MapBlocksToResponse(queries []structs.Query, blocksResp structs.QueriesResp) ([]byte, error) {
	var resp MapSlice
	var err error

	resp = make([]MapItem, len(queries))
	for _, query := range queries {
		blocks, ok := blocksResp[query.Name]
		if !ok {
			return nil, errors.New("Response is empty")
		}

		bLen := len(blocks)
		row := MapSlice{}
		rows := make([]MapSlice, bLen)
		i := 0
		for _, bAndTxs := range blocks {
			if row, err = fieldsResp(&query, bAndTxs); err != nil {
				return nil, err
			}

			rows[i] = row
			i++
		}

		if bLen == 1 {
			resp[query.Order] = MapItem{
				Key:   query.Name,
				Value: row,
			}
		} else {
			resp[query.Order] = MapItem{
				Key:   query.Name,
				Value: rows,
			}
		}
	}

	return resp.MarshalJSON()
}

func fieldsResp(q *structs.Query, blockAndTx structs.BlockAndTx) (resp MapSlice, err error) {
	// resp = make([]MapItem, len(q.Fields))

	qStr := strings.ToLower(q.Name)
	parseBlock := strings.Contains(qStr, "block")
	parseTransaction := strings.Contains(qStr, "transaction")

	if parseBlock && !parseTransaction {
		resp = mapStructToFields(q.Fields, blockAndTx.Block)
	} else if !parseBlock && parseTransaction {
		resp = []MapItem{{
			Key:   q.Name,
			Value: mapSliceToFields(q.Fields, blockAndTx.Txs),
		}}
	} else {
		resp = mapStructToFields(q.Fields, blockAndTx)

	}

	return resp, err
}

func mapStructToFields(fields map[string]structs.Field, s interface{}) MapSlice {
	var value interface{}
	v := reflect.Indirect(reflect.ValueOf(s))
	respMap := make(map[int]MapItem)

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

		fieldType := reflect.TypeOf(v.Field(i).Interface())
		fieldKind := fieldType.Kind()

		switch fieldType {
		case reflect.TypeOf(time.Time{}):
			value = formatValue(v.Field(i).Interface())
		default:
			switch fieldKind {
			case reflect.Slice:
				value = mapSliceToFields(field.Fields, v.Field(i).Interface())
			case reflect.Struct:
				value = mapStructToFields(field.Fields, v.Field(i).Interface())
			default:
				value = formatValue(v.Field(i).Interface())
			}
		}

		respMap[field.Order] = MapItem{
			Key:   field.Name,
			Value: value,
		}
	}

	ms := make([]MapItem, len(respMap))
	i := 0
	for _, resp := range respMap {
		ms[i] = resp
		i++
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
		sliceResp[i] = mapStructToFields(fields, v.Index(i).Interface())
	}

	return sliceResp
}

type MapItem struct {
	Key, Value interface{}
}

type MapSlice []MapItem

func (ms MapSlice) MarshalJSON() ([]byte, error) {
	var b []byte
	var err error
	buf := &bytes.Buffer{}

	buf.Write([]byte{'{'})

	for i, mi := range ms {

		switch reflect.ValueOf(mi.Value) {
		case reflect.ValueOf([]MapSlice{}):
			buf.Write([]byte{'['})
			for i, ms := range mi.Value.([]MapSlice) {
				b, err = ms.MarshalJSON()
				if err != nil {
					return nil, err
				}
				buf.Write(b)
				if i < len(ms)-1 {
					buf.Write([]byte{','})
				}
			}
			buf.Write([]byte{']'})

		case reflect.ValueOf(MapSlice{}):
			b, err = mi.Value.(MapSlice).MarshalJSON()
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
