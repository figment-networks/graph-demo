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

func (s *Service) ProcessGraphqlQuery(ctx context.Context, q []byte, v map[string]interface{}) ([]byte, error) {
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("error while parsing graphql query: %w", err)
	}

	recordsMap := make(map[int][]map[string]interface{})

	for _, query := range queries.Queries {
		// key, ok := query.Params["key"]
		// if !ok {
		// 	return nil, errors.New("missing required part: key")
		// }

		// keyParam, ok := key.Params["key"]
		// if !ok {
		// 	return nil, errors.New("missing required parameter: id")
		// }

		// value, ok := query.Params["value"]
		// if !ok {
		// 	return nil, errors.New("missing required part: value")
		// }

		// valueParam, ok := value.Params["value"]
		// if !ok {
		// 	return nil, errors.New("missing required parameter: value")
		// }

		// if keyParam.Variable != "string" || valueParam.Variable != "string" {
		// 	return nil, errors.New("unexpected parameter variable")
		// }

		// var queryTransactions bool

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

		_, queryBlock := query.Fields["block"]
		if queryBlock {

			recordsMap[query.Order], err = s.store.Get(ctx, "simple-example", "Block", "height", heightValue)
			if err != nil {
				return nil, err
			}

			// if _, queryTransactions = blockField.Fields["transactions"]; queryTransactions {
			// 	recordsMap[query.Order], err = s.store.Get(ctx, "name", "Transaction", id.Name, idParam.Value.(string))
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// }

		} else {
			// if _, queryTransactions = query.Fields["transaction"]; !queryTransactions {
			// 	return nil, errors.New("query has no fields to map")
			// }
		}

		// if queryTransactions {

		// }

	}

	/*
		queries, err := graphcall.ParseQuery(q, v)
		if err != nil {
			return nil, fmt.Errorf("error while parsing graphql query: %w", err)
		}

			for _, query := range queries.Queries {
				s.store.Get(ctx, query.Name, query.)
			}
	*/

	return mapRecordsToResponse(queries.Queries, recordsMap) // s.client.ProcessGraphqlQuery(ctx, q, v)
}

func mapRecordsToResponse(queries []graphcall.Query, recordsMap map[int][]map[string]interface{}) ([]byte, error) {
	var resp qStructs.MapSlice
	// var err error

	resp = make([]qStructs.MapItem, len(queries))
	for _, query := range queries {
		blockRecords, ok := recordsMap[query.Order]
		if !ok {
			continue
		}

		for name, fields := range query.Fields {

			bLen := len(blockRecords)
			var response interface{}
			responses := make([]interface{}, bLen)

			for i, record := range blockRecords {
				response = make(qStructs.MapSlice, len(fields.Fields))
				for _, field := range fields.Fields {
					record, ok := record[field.Name]
					if !ok {
						return nil, errors.New("unknown field name")
					}

					response.(qStructs.MapSlice)[field.Order] = qStructs.MapItem{
						Key:   field.Name,
						Value: record,
					}
				}

				responses[i] = response
			}

			if bLen > 1 {
				response = responses
			}

			resp[query.Order] = qStructs.MapItem{
				Key: query.Name,
				Value: qStructs.MapItem{
					Key:   name,
					Value: response,
				},
			}
		}

	}

	return resp.MarshalJSON()
}
