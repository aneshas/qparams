// Package qparams provides URL query parameter parsing to user defined
// go structs, and provides convenience methods for basic converting of
// parsed parameters (string, int, float64)
package qparams

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type (
	// Map represents qparams map[string]string type
	Map map[string]string

	// Slice represents qparams []string type
	Slice []string
)

// Slice returns string slice of qparams Slice
func (s *Slice) Slice() []string {
	return []string(*s)
}

// ToIntSlice will attempt to convert the string slice to int slice
// Will return error if it is unable to convert any member, and a
// partial slice without errornous members
func (s *Slice) ToIntSlice() ([]int, error) {
	var err error

	newSlice := []int{}

	for _, v := range s.Slice() {
		i, e := strconv.Atoi(v)
		if e != nil {
			err = fmt.Errorf("Could not convert member %s to int", v)
			continue
		}

		newSlice = append(newSlice, i)
	}

	return newSlice, err
}

// ToFloatSlice will attempt to convert the string slice to float64 slice
// Will return error if it is unable to convert any member, and a
// partial slice without errornous members
func (s *Slice) ToFloatSlice() ([]float64, error) {
	var err error

	newSlice := []float64{}

	for _, v := range s.Slice() {
		f, e := strconv.ParseFloat(v, 64)
		if e != nil {
			err = fmt.Errorf("Could not convert member %s to float", v)
			continue
		}

		newSlice = append(newSlice, f)
	}

	return newSlice, err
}

// ToIntAtIndex will attempt to convert i-th member of slice to integer
// Will return conversion error on failure and an int zero value
/*
func (s *Slice) ToIntAtIndex(i int) (int, error) {
	return 0, nil
}

// ToFloatAtIndex will attempt to convert i-th member of slice to float64
// Will return conversion error on failure and an int zero value
func (s *Slice) ToFloatAtIndex(i int) (float64, error) {
	return 0.0, nil
}
*/

// TODO - Add tofloat etc... and do the same for Map

// ErrWrongDestType is used when the provided dest is not struct pointer
var ErrWrongDestType = errors.New("Dest must be a struct pointer")

// TypeConvErrors contain errors generated upon conversion to int or float64
type TypeConvErrors []string

func (e TypeConvErrors) Error() string {
	str := ""

	for _, e := range e {
		str += fmt.Sprintf("%s\n", e)
	}

	return str
}

var separator = ","
var mapOpsTagSeparator = ","

// Parse will try to parse query params from http.Request to
// provided struct, and will return error on filure
func Parse(dest interface{}, r *http.Request) error {
	var errs TypeConvErrors

	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)
	queryValues := r.URL.Query()

	if t.Kind() != reflect.Ptr ||
		t.Elem().Kind() != reflect.Struct {
		return ErrWrongDestType
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

		switch fieldV.Kind() {
		case reflect.Int:
			err := parseInt(fieldT, fieldV, queryValue)
			if err != nil {
				errs = append(errs, err.Error())
			}
		case reflect.Float64:
			err := parseFloat64(fieldT, fieldV, queryValue)
			if err != nil {
				errs = append(errs, err.Error())
			}
		case reflect.String:
			parseString(fieldT, fieldV, queryValue)
		}
	}

	if errs != nil && len(errs) > 0 {
		return errs
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

func getOperators(sField reflect.StructField) []string {
	operators := []string{}

	if ops := getTag("ops", sField); ops != "" {
		operators = strings.Split(ops, mapOpsTagSeparator)
	}

	return operators
}

func parseMap(sField reflect.StructField, fieldV reflect.Value, queryValue string) {
	sep := getSeparator(sField)

	operators := getOperators(sField)
	// TODO - Throw error if no operators provided

	// TODO - handle error
	parsedMap := walk(queryValue, sep, operators)

	fieldV.Set(reflect.ValueOf(parsedMap))
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

func parseInt(sField reflect.StructField, fieldV reflect.Value, queryValue string) error {
	i, err := strconv.Atoi(queryValue)
	if err != nil {
		return fmt.Errorf("Field %s does not contain a valid integer (%s)", sField.Name, queryValue)
	}

	fieldV.Set(reflect.ValueOf(i))

	return nil
}

func parseFloat64(sField reflect.StructField, fieldV reflect.Value, queryValue string) error {
	f, err := strconv.ParseFloat(queryValue, 64)
	if err != nil {
		return fmt.Errorf("Field %s does not contain a valid float (%s)", sField.Name, queryValue)
	}

	fieldV.Set(reflect.ValueOf(f))

	return nil
}

func parseString(sField reflect.StructField, fieldV reflect.Value, queryValue string) {
	fieldV.Set(reflect.ValueOf(queryValue))
}
