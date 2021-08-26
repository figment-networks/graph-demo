package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/figment-networks/graph-demo/graphcall"
	qStructs "github.com/figment-networks/graph-demo/graphcall/response"
	"github.com/figment-networks/graph-demo/runner/store/memap"
)

type Service struct {
	store *memap.SubgraphStore
}

func New(store *memap.SubgraphStore) *Service {
	return &Service{
		store: store,
	}
}

type qRecordsMap map[int][]map[string]interface{}

func (s *Service) ProcessGraphqlQuery(ctx context.Context, subgraph string, q []byte, v map[string]interface{}) ([]byte, error) {
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("error while parsing graphql query: %w", err)
	}

	recordsMap := make(qRecordsMap)

	for _, query := range queries.Queries {
		for n, p := range query.Params {
			sVal, err := getStringParam(n, p)
			if err != nil {
				return nil, err
			}
			records, err := s.store.Get(ctx, subgraph, query.Name, n, sVal)
			if err != nil && err != memap.ErrRecordsNotFound {
				return nil, err
			}
			recordsMap[query.Order] = records
			break
		}
	}

	return mapRecordsToResponse(queries.Queries, recordsMap)
}

func getStringParam(name string, param graphcall.Part) (str string, err error) {

	pParam, ok := param.Params[name]
	if !ok {
		return "", errors.New("missing required parameter: " + name)
	}

	switch pParam.Variable {
	case "string":
		str = pParam.Value.(string)
	case "uint64":
		hUint64 := pParam.Value.(uint64)
		str = strconv.Itoa(int(hUint64))
	case "float64":
		hF64 := pParam.Value.(float64)
		str = strconv.FormatFloat(hF64, 'E', -1, 64)
	default:
		return "", fmt.Errorf("unexpected parameter variable %q", pParam.Variable)
	}

	return str, nil
}

func mapRecordsToResponse(queries []graphcall.Query, recordsMap qRecordsMap) ([]byte, error) {
	var response interface{}
	var resp qStructs.MapSlice
	var err error

	resp = make([]qStructs.MapItem, len(queries))
	for _, query := range queries {
		response, err = mapBlockAndTxsToResponse(recordsMap[query.Order], query.Fields)
		if err != nil {
			return nil, err
		}

		resp[query.Order] = qStructs.MapItem{
			Key:   query.Name,
			Value: response,
		}

	}

	return resp.MarshalJSON()
}

func mapBlockAndTxsToResponse(records []map[string]interface{}, fields map[string]graphcall.Field) (interface{}, error) {
	var response interface{}
	var err error
	rLen := len(records)
	responses := make([]interface{}, rLen)

	for i, record := range records {
		if response, err = fieldsStructResponse(fields, record); err != nil {
			return nil, err
		}

		responses[i] = response
	}

	if rLen > 1 {
		response = responses
	}

	return response, nil
}

func fieldsStructResponse(fields map[string]graphcall.Field, record map[string]interface{}) (qStructs.MapSlice, error) {
	var err error
	response := make(map[int]qStructs.MapItem, len(fields))
	maxOrder := 0

	for _, field := range fields {
		recordValue, ok := record[field.Name]
		if !ok {
			return nil, fmt.Errorf("unknown field name %q", field.Name)
		}

		var value interface{}
		if field.Fields != nil {
			txs := recordValue.([]map[string]interface{})
			txsMs := make([]qStructs.MapSlice, len(txs))
			for i, tx := range txs {
				txsMs[i], err = fieldsStructResponse(field.Fields, tx)
				if err != nil {
					return nil, err
				}
			}
			value = txsMs
		} else {
			value = recordValue
		}

		if maxOrder < field.Order {
			maxOrder = field.Order
		}

		response[field.Order] = qStructs.MapItem{
			Key:   field.Name,
			Value: value,
		}
	}

	respLen := len(response)
	ms := make([]qStructs.MapItem, respLen)

	i := 0
	for order := 0; order <= maxOrder; order++ {
		resp, ok := response[order]
		if ok && resp.Key != nil {
			ms[i] = resp
			i++
		}
	}

	if i == 0 {
		return nil, nil
	}

	return ms, nil
}
