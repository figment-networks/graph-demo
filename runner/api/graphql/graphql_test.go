package graphql

import (
	"reflect"
	"testing"
)

const t1 = `input Sth {
	str: String
}

input ReviewInput {
	stars: Int!
	commentary: [String]
	additional: Sth
  }
  
  query GetBlock($height: Int, $review: ReviewInput) {
    secondBlock(sHeight: $height) {
		height
		hash
	  }
	
	block(height: $height ) {
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
			query:     t1,
			variables: map[string]interface{}{"height": int32(120), "review": map[string]interface{}{"stars": int32(3), "commentary": []string{"a", "b"}, "b": 2, "additional": map[string]interface{}{"str": "Hello it's me"}}},
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
