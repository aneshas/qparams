// Package qparams provides custom URL query parameter parsing
package qparams

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type (
	Map   map[string]string
	Slice []string
)

func (s *Slice) ToIntSlice() []int {
	return nil
}

func (s *Slice) ToIntAtIndex(i int) int {
	return 0
}

// TODO - Add tofloat etc... and do the same for Map

var DestTypeError = errors.New("Dest must be a struct pointer")

var separator = ","

// Parse will try to parse query params from http.Request to
// provided struct, and will return error on filure
func Parse(dest interface{}, r *http.Request) error {
	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)
	queryValues := r.URL.Query()

	if t.Kind() != reflect.Ptr &&
		t.Elem().Kind() != reflect.Struct {
		return DestTypeError
	}

	// TODO - Cache struct meta data

	for i := 0; i < v.Elem().NumField(); i++ {
		fieldT := t.Elem().Field(i)
		fieldV := v.Elem().Field(i)

		fieldName := strings.ToLower(fieldT.Name)

		for key, val := range queryValues {
			key = strings.ToLower(key)
			queryValues[key] = val
		}

		queryValue := queryValues.Get(fieldName)

		if queryValue == "" {
			// TODO - Set default value here
			continue
		}

		switch fieldT.Type.Name() {
		case "Map":
			parseMap(fieldT, fieldV, queryValue)
		case "Slice":
			parseSlice(fieldT, fieldV, queryValue)
		}
	}

	return nil
}

func getTag(tag string, sField reflect.StructField) string {
	tags := sField.Tag.Get("qparams")

	if tags == "" {
		return tags
	}

	tagSlice := strings.Split(tags, " ")

	for _, t := range tagSlice {
		subSlice := strings.Split(t, ":")

		if subSlice != nil &&
			len(subSlice) == 2 &&
			subSlice[0] == tag {
			return subSlice[1]
		}
	}

	return ""
}

func getSeparator(sField reflect.StructField) string {
	sep := separator

	if s := getTag("sep", sField); s != "" {
		sep = s
	}

	return sep
}

func parseMap(sField reflect.StructField, fieldV reflect.Value, queryValue string) {
	sep := getSeparator(sField)
	fmt.Printf("Map Separator %s\n", sep)
}

func parseSlice(sField reflect.StructField, fieldV reflect.Value, queryValue string) {
	sep := getSeparator(sField)

	slice := strings.Split(queryValue, sep)

	newSlice := Slice{}

	for _, val := range slice {
		v := strings.ToLower(val)
		if v != "" {
			newSlice = append(newSlice, v)
		}
	}

	fieldV.Set(reflect.ValueOf(newSlice))
}
