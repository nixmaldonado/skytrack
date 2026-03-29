package model

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

// ICAOCode is a 4-letter ICAO airport code (e.g., "KJFK", "EGLL").
type ICAOCode string

var icaoRegex = regexp.MustCompile(`^[A-Z]{4}$`)

func (c ICAOCode) IsValid() bool {
	return icaoRegex.MatchString(string(c))
}

func MarshalICAOCode(c ICAOCode) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, `"`+string(c)+`"`)
	})
}

func UnmarshalICAOCode(v interface{}) (ICAOCode, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("ICAOCode must be a string")
	}
	s = strings.ToUpper(strings.TrimSpace(s))
	code := ICAOCode(s)
	if !code.IsValid() {
		return "", fmt.Errorf("ICAOCode must be exactly 4 uppercase letters, got %q", s)
	}
	return code, nil
}

// IATACode is a 3-letter IATA airport code (e.g., "JFK", "LHR").
type IATACode string

var iataRegex = regexp.MustCompile(`^[A-Z]{3}$`)

func (c IATACode) IsValid() bool {
	return iataRegex.MatchString(string(c))
}

func MarshalIATACode(c IATACode) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, `"`+string(c)+`"`)
	})
}

func UnmarshalIATACode(v interface{}) (IATACode, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("IATACode must be a string")
	}
	s = strings.ToUpper(strings.TrimSpace(s))
	code := IATACode(s)
	if !code.IsValid() {
		return "", fmt.Errorf("IATACode must be exactly 3 uppercase letters, got %q", s)
	}
	return code, nil
}
