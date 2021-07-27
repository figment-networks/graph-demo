package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	runnerHTTP "github.com/figment-networks/graph-demo/manager/api/runner/transport/http"

	"go.uber.org/zap"
)

type Service struct {
	url string
	cli *http.Client
	log *zap.Logger
}

func New(cli *http.Client, log *zap.Logger, url string) *Service {
	return &Service{
		url: url,
		cli: cli,
		log: log,
	}
}

func (s *Service) ProcessGraphqlQuery(ctx context.Context, v map[string]interface{}, q string) ([]byte, error) {
	s.log.Debug("[HTTP] Sending process graphql query request")

	buff := new(bytes.Buffer)
	enc := json.NewEncoder(buff)
	reqBody := runnerHTTP.JSONGraphQLRequest{
		Query:     q,
		Variables: v,
	}

	if err := enc.Encode(reqBody); err != nil {
		s.log.Error("Error while encoding request", zap.Error(err))
		return nil, err
	}

	body := bytes.NewReader(buff.Bytes())

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/graphQL", s.url), body)
	if err != nil {
		s.log.Error("[HTTP] Error while creating a new request", zap.Error(err))
		return nil, err
	}

	resp, err := s.cli.Do(req)
	if err != nil {
		s.log.Error("[HTTP] Error while getting response from manager", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	byteResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.log.Error("[HTTP] Error while reading graphql response body", zap.Error(err))
		return nil, err
	}

	s.log.Debug("[HTTP] Received process graphql query response")
	return byteResp, nil
}
