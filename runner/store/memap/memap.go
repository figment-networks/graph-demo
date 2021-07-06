package memap

import (
	"context"
	"errors"
	"fmt"

	"github.com/figment-networks/graph-demo/runner/store"
)

type Stor struct {
	ID              string
	IndexedFields   []string
	ReferenceFields []store.NT
	Records         map[string]*Record

	Indexes map[string]map[string][]*Record
}

type Record struct {
	Data      map[string]interface{}
	Reference map[string][]store.NT
}

type SubgraphStore struct {
	s map[string]*MemoryMapStore // subgraph name
}

func NewSubgraphStore() *SubgraphStore {
	return &SubgraphStore{
		s: make(map[string]*MemoryMapStore),
	}
}

type MemoryMapStore struct {
	storages map[string]Stor // structure name
}

func NewMemoryMapStore() *MemoryMapStore {
	return &MemoryMapStore{
		storages: make(map[string]Stor),
	}
}

func (ss *SubgraphStore) NewStore(name, structure string, indexed []store.NT) {
	s := Stor{
		Records: make(map[string]*Record),
		Indexes: make(map[string]map[string][]*Record),
	}

	for _, nt := range indexed {
		if nt.IsArray {
			s.ReferenceFields = append(s.ReferenceFields, nt)
			continue
		}

		switch nt.Type {
		case "ID":
			s.ID = nt.Name
			s.IndexedFields = append(s.IndexedFields, nt.Name)
			s.Indexes[nt.Name] = make(map[string][]*Record)
		case "String":
			s.IndexedFields = append(s.IndexedFields, nt.Name)
			s.Indexes[nt.Name] = make(map[string][]*Record)
		}
	}

	mms, ok := ss.s[name]
	if !ok {
		mms = NewMemoryMapStore()
	}

	mms.storages[structure] = s
	ss.s[name] = mms
}

func (ss *SubgraphStore) Store(ctx context.Context, name, structure string, data map[string]interface{}) error {
	subgraph, ok := ss.s[name]
	if !ok {
		return fmt.Errorf("subgraph not found")
	}

	return subgraph.Store(ctx, structure, data)
}

func (ss *SubgraphStore) Get(ctx context.Context, name, kind, key, value string) (records []map[string]interface{}, err error) {
	subgraph, ok := ss.s[name]
	if !ok {
		return nil, fmt.Errorf("subgraph not found")
	}

	return subgraph.Get(ctx, kind, key, value)
}

func (mm *MemoryMapStore) Store(ctx context.Context, name string, data map[string]interface{}) error {

	s, ok := mm.storages[name]
	if !ok {
		return errors.New("storage does not exists")
	}

	r := &Record{Data: data}
	id, ok := data[s.ID]
	if !ok {
		return errors.New("primary key not present")
	}
	idS, ok := id.(string)
	if !ok {
		return errors.New("primary key is not a string")
	}
	s.Records[idS] = r

	for _, in := range s.IndexedFields {
		val, ok := data[in]
		if !ok {
			return fmt.Errorf("expected field %s not present", in)
		}

		sval, ok := val.(string)
		if !ok {
			return fmt.Errorf("expected field %s is not a string: %+v", in, val)
		}

		k, ok := s.Indexes[in]
		if !ok {
			return fmt.Errorf("index not found")
		}

		keys, ok := k[sval]
		if !ok {
			keys = []*Record{}
		}
		keys = append(keys, r)
		k[sval] = keys
		s.Indexes[in] = k
	}

	return nil
}

func (mm *MemoryMapStore) Get(ctx context.Context, kind, key, value string) (records []map[string]interface{}, err error) {
	k := mm.storages[kind]
	record, ok := k.Indexes[key]
	if !ok {
		return nil, fmt.Errorf(" (%s) is not an indexed field", key)
	}

	found, ok := record[value]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	for _, f := range found {
		records = append(records, f.Data)
	}

	return records, nil
}
