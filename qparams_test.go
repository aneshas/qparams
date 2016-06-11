package qparams

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

const failEmoji = "\x1b[31m\u2717\x1b[0m"
const passEmoji = "\x1b[92m\u2713\x1b[0m"

func failFatal(t *testing.T,
	msg string,
	want, got interface{},
	args ...interface{}) {

	message := fmt.Sprintf("\t%s %s! WANT: %+v GOT: %+v", failEmoji, msg, want, got)
	t.Fatalf(message, args...)
}

func pass(t *testing.T, msg string, want, got interface{}) {
	t.Logf("\t%s %s! WANT: %+v GOT: %+v", passEmoji, msg, want, got)
}

type testCase struct {
	URL            string
	ExpectedResult interface{}
	ExpectedError  error
}

func newRequest(url string) *http.Request {
	r, _ := http.NewRequest("GET", url, nil)
	return r
}

func compare(t *testing.T, c testCase, got interface{}, err error) {
	if err != c.ExpectedError {
		failFatal(t, "Incorrect error value", nil, err)
	}

	switch reflect.DeepEqual(got, c.ExpectedResult) {
	case true:
		pass(t, "Test passed", c.ExpectedResult, got)
	case false:
		failFatal(t, "Test failed", c.ExpectedResult, got)
	}
}

// MARK - Tests

func TestParseSlice(t *testing.T) {
	type testStruct struct {
		Embed Slice
	}

	table := []testCase{
		{
			URL:            "foobar.com?embed=User,Order,Discount",
			ExpectedResult: testStruct{Embed: Slice{"user", "order", "discount"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?Embed=User,Order,Discount",
			ExpectedResult: testStruct{Embed: Slice{"user", "order", "discount"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?embed=User,Order,",
			ExpectedResult: testStruct{Embed: Slice{"user", "order"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?Embed=,User,Order,",
			ExpectedResult: testStruct{Embed: Slice{"user", "order"}},
			ExpectedError:  nil,
		},
	}

	t.Log("")
	t.Log("Testing slice parsing")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)
	}
}

func TestParseSliceCustomSeparator(t *testing.T) {
	type testStruct struct {
		Embed Slice `qparams:"sep:|"`
	}

	table := []testCase{
		{
			URL:            "foobar.com?embed=User|Order|Discount",
			ExpectedResult: testStruct{Embed: Slice{"user", "order", "discount"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?Embed=User|Order|Discount",
			ExpectedResult: testStruct{Embed: Slice{"user", "order", "discount"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?embed=User|Order|",
			ExpectedResult: testStruct{Embed: Slice{"user", "order"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?Embed=|User|Order|",
			ExpectedResult: testStruct{Embed: Slice{"user", "order"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?Embed=",
			ExpectedResult: testStruct{},
			ExpectedError:  nil,
		},
	}

	t.Log("")
	t.Log("Testing slice parsing with custom separator")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)
	}
}
