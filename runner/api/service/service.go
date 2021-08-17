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

type qRecordsMap map[int][]map[string]interface{}

func (s *Service) ProcessGraphqlQuery(ctx context.Context, q []byte, v map[string]interface{}) ([]byte, error) {
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("error while parsing graphql query: %w", err)
	}

	recordsMap := make(qRecordsMap)

	for _, query := range queries.Queries {

		hFloat, heightStr, err := getHeight(query.Params)
		if err != nil {
			return nil, err
		}

		var queryTransactions bool
		blockField, queryBlock := query.Fields["block"]
		if queryBlock {

			records, err := s.store.Get(ctx, "simple-example", "Block", "height", heightStr)
			if err != nil && err != memap.ErrRecordsNotFound {
				return nil, err
			}
			recordsMap[query.Order] = records

			_, queryTransactions = blockField.Fields["transactions"]

		} else {
			if _, queryTransactions = query.Fields["transactions"]; !queryTransactions {
				return nil, errors.New("query has no fields to map")
			}
		}

		if queryTransactions {
			records, err := s.store.Get(ctx, "simple-example", "Transaction", "height", heightStr)
			if err != nil && err != memap.ErrRecordsNotFound {
				return nil, err
			}

			if queryBlock {
				for i, blockRecords := range recordsMap[query.Order] {
					heightRecord, ok := blockRecords["height"]
					if ok && heightRecord == hFloat {
						recordsMap[query.Order][i]["transactions"] = records
					}
				}
			} else {
				recordsMap[query.Order] = records
			}
		}
	}

	return mapRecordsToResponse(queries.Queries, recordsMap) // s.client.ProcessGraphqlQuery(ctx, q, v)
}

func getHeight(params map[string]graphcall.Part) (hFloat float64, heightStr string, err error) {
	heightPart, ok := params["height"]
	if !ok {
		return 0, "", errors.New("missing required part: height")
	}

	heightParam, ok := heightPart.Params["height"]
	if !ok {
		return 0, "", errors.New("missing required parameter: height")
	}

	switch heightParam.Variable {
	case "string":
		heightStr = heightParam.Value.(string)
		hInt, err := strconv.Atoi(heightStr)
		if err != nil {
			return 0, "", err
		}
		hFloat = float64(hInt)
	case "uint64":
		hUint64 := heightParam.Value.(uint64)
		heightStr = strconv.Itoa(int(hUint64))
		hFloat = float64(hUint64)
	default:
		return 0, "", fmt.Errorf("unexpected parameter variable %q", heightParam.Variable)
	}

	return hFloat, heightStr, nil
}

func mapRecordsToResponse(queries []graphcall.Query, recordsMap qRecordsMap) ([]byte, error) {
	var response interface{}
	var resp qStructs.MapSlice
	var err error

	resp = make([]qStructs.MapItem, len(queries))
	for _, query := range queries {

		for name, fields := range query.Fields {
			records, _ := recordsMap[query.Order]
			response, err = mapBlockAndTxsToResponse(records, fields.Fields)
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

func mapBlockAndTxsToResponse(records []map[string]interface{}, fields map[string]graphcall.Field) (interface{}, error) {
	var response interface{}
	var err error
	rLen := len(records)
	responses := make([]interface{}, rLen)

	// if records == nil {
	// 	return nil, nil
	// }

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
			txs := recordValue.(interface{}).([]map[string]interface{})
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
