package graphql

import (
	"errors"
	"log"
	"regexp"
	"strings"
)

var (
	partRegxp = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
	params    = regexp.MustCompile("\\s*([a-zA-Z0-9_-]+)\\s*(|\\(?[a-zA-Z0-9\\=\\,\\s\\.\\$\\_\\-\\:\"\\!]*\\))\\s*({?)\\n")
)

type Query struct {
	Q      Part
	Params map[string]Part
}

type Part struct {
	Name   string
	Params map[string]interface{}
}

func ParseQuery(query string, variables map[string]interface{}) (q Query, err error) {
	q = Query{}
	pos := strings.Index(query, "{")
	if pos < 0 {
		return q, errors.New("Invalid grapghql")
	}

	initialPortion := query[0 : pos+1]
	if strings.Contains(initialPortion, "query") {
		parts := partRegxp.FindAllStringSubmatch(initialPortion+"\n", -1)
		log.Println("parts", parts)
		if len(parts) == 1 {
			q.Q = Part{
				Name:   parts[0][1],
				Params: parseParams(parts[0][2]),
			}
		}
	}

	return q, nil
}

func parseParams(paramS string) map[string]interface{} {

}
