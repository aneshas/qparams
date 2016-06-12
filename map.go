package qparams

import (
	"fmt"
	"strings"
	"unicode"
)

func upercaseInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func isOperator(c string, operators []string) (bool, int) {

	for _, o := range operators {
		if o == c {
			return true, len(c)
		}
	}

	if len(c) >= 4 {
		single := c[0:4]
		for _, o := range operators {
			if o == single {
				return true, 4
			}
		}
	}

	if len(c) >= 2 {
		single := c[0:2]
		for _, o := range operators {
			if o == single {
				return true, 2
			}
		}
	}

	single := c[0:1]
	for _, o := range operators {
		if o == single {
			return true, 1
		}
	}

	return false, 0
}

func getValue(str, separator string) (string, int) {
	var i int
	var chunk string
	var c rune

	for i, c = range str {
		if c != rune(separator[0]) {
			chunk += string(c)
			continue
		}
	}

	return chunk, i
}

func walk(filterRaw string, separator string, operators []string) map[string]string {
	filters := make(Map)

	strSlice := strings.Split(filterRaw, separator)
	for _, filter := range strSlice {
		if filter == "" {
			continue
		}

		var chunk string
		var op string

		for i, c := range filter {
			var off int
			var cmp string

			if i < len(filter)-6 {
				cmp = filter[i : i+6]
			} else {
				cmp = filter[i:]
			}

			if isO, count := isOperator(cmp, operators); isO {
				op = string(c)
				off += count
				if count >= 2 { //&& c != '=' {
					op += string(filter[i+1 : i+count])
				}
			} else {
				chunk += string(c)
				continue
			}

			value, _ := getValue(filter[i+off:], separator)
			key := fmt.Sprintf("%s %s", strings.ToLower(chunk), op)

			filters[key] = value
		}
	}

	return filters
}
