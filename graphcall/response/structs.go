package structs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

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
		switch reflect.ValueOf(mi.Value).Type().String() {
		case reflect.ValueOf([]MapSlice{}).Type().String():
			sliceValue := mi.Value.([]MapSlice)
			sliceLen := len(sliceValue)
			sliceBytes := []byte{'['}
			for i, ms := range sliceValue {
				valueBytes, err := ms.MarshalJSON()
				if err != nil {
					return nil, err
				}
				sliceBytes = append(sliceBytes, valueBytes...)
				if i < sliceLen-1 {
					sliceBytes = append(sliceBytes, ',')
				}
			}
			b = append(sliceBytes, ']')

		case reflect.ValueOf(MapSlice{}).Type().String():
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
