package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

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

	resp, err := http.Get(fmt.Sprintf("%s/graphQL", s.url))
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
