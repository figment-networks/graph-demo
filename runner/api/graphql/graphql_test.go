package graphql

import (
	"reflect"
	"testing"
)

const t1 = `query GetBlock($height: Int) {
    block( $height: Int = 0 ) {
      height
      time
      id
    }
  }`

func TestParseQuery(t *testing.T) {
	type args struct {
		query     string
		variables map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    Query
		wantErr bool
	}{
		{name: "simple", args: args{
			query: t1,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuery(tt.args.query, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
