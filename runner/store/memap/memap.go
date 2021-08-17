package memap

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/figment-networks/graph-demo/runner/store"
)

var (
	ErrRecordsNotFound  = fmt.Errorf("not found")
	ErrSubgraphNotFound = fmt.Errorf("subgraph not found")
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
		case "Int":
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

func (ss *SubgraphStore) Store(ctx context.Context, data map[string]interface{}, name, structure string) error {
	subgraph, ok := ss.s[name]
	if !ok {
		return ErrSubgraphNotFound
	}
	return subgraph.Store(ctx, data, structure)
}

func (ss *SubgraphStore) Get(ctx context.Context, name, structure, key, value string) (records []map[string]interface{}, err error) {
	subgraph, ok := ss.s[name]
	if !ok {
		return nil, ErrSubgraphNotFound
	}
	return subgraph.Get(ctx, structure, key, value)
}

// map[height]map[Block/Transaction][]*Records
func (mm *MemoryMapStore) Store(ctx context.Context, data map[string]interface{}, structure string) error {
	s, ok := mm.storages[structure]
	if !ok {
		return fmt.Errorf("storage does not exists for structure %q", structure)
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

		var stringValue string
		fieldType := reflect.ValueOf(val).Kind()
		switch fieldType {
		case reflect.Float64:
			var fVal float64
			fVal, ok = val.(float64)
			stringValue = strconv.Itoa(int(fVal))
		case reflect.String:
			stringValue, ok = val.(string)
		default:
			return fmt.Errorf("unexpected field type %s %s: %+v", fieldType, in, val)
		}

		if !ok {
			return fmt.Errorf("could not read field value as %s: %s: %+v", fieldType, in, val)
		}

		k, ok := s.Indexes[in]
		if !ok {
			return fmt.Errorf("index not found")
		}

		keys, ok := k[stringValue]
		if !ok {
			keys = []*Record{}
		}
		keys = append(keys, r)
		k[stringValue] = keys
		s.Indexes[in] = k
	}

	return nil
}

func (mm *MemoryMapStore) Get(ctx context.Context, structure, key, value string) (records []map[string]interface{}, err error) {
	k := mm.storages[structure]
	record, ok := k.Indexes[key]
	if !ok {
		return nil, fmt.Errorf(" (%s) is not an indexed field", key)
	}

	found, ok := record[value]
	if !ok {
		return nil, ErrRecordsNotFound
	}

	for _, f := range found {
		records = append(records, f.Data)
	}

	return records, nil
}
