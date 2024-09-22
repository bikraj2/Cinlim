package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFromat = errors.New("the provided string is no")

type Runtime int32

func (rt Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", rt)
	quoteJsonValue := strconv.Quote(jsonValue)
	return []byte(quoteJsonValue), nil
}
func (rt *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unquottedString, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFromat
	}

	parts := strings.Split(unquottedString, " ")
	// Check for sanity of the thing
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFromat
	}

	i, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return ErrInvalidRuntimeFromat
	}
	*rt = Runtime(i)
	return nil
}
