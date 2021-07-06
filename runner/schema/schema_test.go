package schema

import (
	"log"
	"testing"
)

func TestSchemas_LoadFromFile(t *testing.T) {
	type args struct {
		name string
		path string
	}
	tests := []struct {
		name    string
		s       *Schemas
		args    args
		wantErr bool
	}{
		{name: "a", args: args{name: "one", path: "../subgraphs/subgraphOne/schema.graphql"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSchemas()
			if err := s.LoadFromFile(tt.args.name, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Schemas.LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			log.Printf("e %+v", s.Subgraphs)

		})
	}
}
