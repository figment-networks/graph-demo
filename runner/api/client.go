package api

import (
	"fmt"

	"github.com/figment-networks/graph-demo/runner/api/graphql"
)

type Client struct {
}

func (c *Client) FetchData(q *graphql.GraphQuery) error {
	for _, q := range q.Queries {
		fmt.Println(q)
	}

	return nil
}
