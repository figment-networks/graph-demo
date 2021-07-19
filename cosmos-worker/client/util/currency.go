package util

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
)

var curencyRegex = regexp.MustCompile("([0-9\\.\\,\\-]+)[\\s]*([^0-9\\s]+)$")

func GetCoin(s string) (number *big.Int, exp int32, err error) {
	s = strings.Replace(s, ",", ".", -1)
	strs := strings.Split(s, `.`)
	number = &big.Int{}
	if len(strs) == 1 {
		number.SetString(strs[0], 10)
		return number, 0, nil
	}
	if len(strs) == 2 {
		number.SetString(strs[0]+strs[1], 10)
		return number, int32(len(strs[1])), nil
	}

	return number, 0, errors.New("Impossible to parse ")
}

func GetCurrency(in string) []string {
	return curencyRegex.FindStringSubmatch(in)
}
