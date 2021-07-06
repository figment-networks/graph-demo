package store

import "context"

type NT struct {
	Name    string
	Type    string
	IsArray bool
}

type Storage interface {
	NewStore(name, structure string, indexed []NT)
	Store(ctx context.Context, name, structure string, data map[string]interface{}) error
	Get(ctx context.Context, name, structure, key, value string) (records []map[string]interface{}, err error)
}
