package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/figment-networks/graph-demo/graphcall"
	qStructs "github.com/figment-networks/graph-demo/graphcall/response"
	"github.com/figment-networks/graph-demo/runner/store/memap"

	"go.uber.org/zap"
)

type Service struct {
	store *memap.SubgraphStore
	log   *zap.Logger
}

func New(store *memap.SubgraphStore) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) ProcessGraphqlQuery(ctx context.Context, subgraph string, q []byte, v map[string]interface{}) ([]byte, error) {
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("error while parsing graphql query: %w", err)
	}

	recordsMap := make(map[int]map[string][]map[string]interface{})

	for _, query := range queries.Queries {

		heightPart, ok := query.Params["height"]
		if !ok {
			return nil, errors.New("missing required part: height")
		}

		heightParam, ok := heightPart.Params["height"]
		if !ok {
			return nil, errors.New("missing required parameter: height")
		}

		var heightValue string
		switch heightParam.Variable {
		case "string":
			heightValue = heightParam.Value.(string)
		case "uint64":
			heightValue = strconv.Itoa(int(heightParam.Value.(uint64)))
		default:
			return nil, fmt.Errorf("unexpected parameter variable %q", heightParam.Variable)
		}

		var queryTransactions bool
		blockField, queryBlock := query.Fields["block"]
		if queryBlock {
			records, err := s.store.Get(ctx, subgraph, "Block", "height", heightValue)
			if err != nil {
				return nil, err
			}

			if _, ok := recordsMap[query.Order]; !ok {
				recordsMap[query.Order] = make(map[string][]map[string]interface{})
			}
			recordsMap[query.Order]["block"] = records

			_, queryTransactions = blockField.Fields["transactions"]

		} else {
			if _, queryTransactions = query.Fields["transactions"]; !queryTransactions {
				return nil, errors.New("query has no fields to map")
			}
		}

		if queryTransactions {
			records, err := s.store.Get(ctx, subgraph, "Transaction", "height", heightValue)
			if err != nil {
				return nil, err
			}

			if _, ok := recordsMap[query.Order]; !ok {
				recordsMap[query.Order] = make(map[string][]map[string]interface{})
			}
			recordsMap[query.Order]["transactions"] = records
		}

	}

	return mapRecordsToResponse(queries.Queries, recordsMap) // s.client.ProcessGraphqlQuery(ctx, q, v)
}

func mapRecordsToResponse(queries []graphcall.Query, recordsMap map[int]map[string][]map[string]interface{}) ([]byte, error) {
	var response interface{}
	var resp qStructs.MapSlice
	var err error

	resp = make([]qStructs.MapItem, len(queries))
	for _, query := range queries {

		for name, fields := range query.Fields {
			txsRecords, txsOk := recordsMap[query.Order]["transactions"]

			switch fields.Name {
			case "block":
				blockRecords, ok := recordsMap[query.Order]["block"]
				if !ok {
					continue
				}
				response, err = mapBlocksToResponse(blockRecords, txsRecords, fields.Fields)
			case "transactions":
				if !txsOk {
					continue
				}
				response, err = mapBlocksToResponse(txsRecords, nil, fields.Fields)
			default:
				return nil, fmt.Errorf("unknown field value to map %q", fields.Name)
			}

			if err != nil {
				return nil, err
			}

			resp[query.Order] = qStructs.MapItem{
				Key: query.Name,
				Value: qStructs.MapSlice{qStructs.MapItem{
					Key:   name,
					Value: response,
				}},
			}
		}

	}

	return resp.MarshalJSON()
}

func mapBlocksToResponse(records, nested []map[string]interface{}, fields map[string]graphcall.Field) (interface{}, error) {
	var response interface{}
	var err error
	nLen := len(nested)
	rLen := len(records)
	responses := make([]interface{}, rLen)

	for i, record := range records {
		var nMap map[string]interface{}
		if i < nLen {
			nMap = nested[i]
		}
		if response, err = fieldsStructResponse(fields, record, nMap); err != nil {
			return nil, err
		}

		responses[i] = response
	}

	if rLen > 1 {
		response = responses
	}

	return response, nil
}

func fieldsStructResponse(fields map[string]graphcall.Field, record, nested map[string]interface{}) (qStructs.MapSlice, error) {
	response := make(map[int]qStructs.MapItem, len(fields))
	maxOrder := 0

	for _, field := range fields {
		var value interface{}
		if field.Name == "transactions" {
			ms, err := fieldsStructResponse(field.Fields, nested, nil)
			if err != nil {
				return nil, err
			}

			value = ms
		} else {
			record, ok := record[field.Name]
			if !ok {
				return nil, errors.New("unknown field name")
			}

			value = record
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
