package store

import "context"

type NT struct {
	Name    string
	Type    string
	IsArray bool
}

type Storage interface {
	NewStore(name, structure string, indexed []NT)
	Store(ctx context.Context, data map[string]interface{}, name, structure string) error
	Get(ctx context.Context, name, structure, key, value string) (records []map[string]interface{}, err error)
}
