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
	URL                      string
	ExpectedResult           interface{}
	ExpectedIntSliceResult   []int
	ExpectedFloatSliceResult []float64
	ExpectedIntResult        int
	ExpectedFloatResult      float64
	ExpectedError            error
	ExpectedConvErr          error
}

func newRequest(url string) *http.Request {
	r, _ := http.NewRequest("GET", url, nil)
	return r
}

func checkErr(t *testing.T, got, want error) {
	if got != nil && want == nil {
		failFatal(t, "Incorrect error value", want, got)
	}

	if got == nil && want != nil {
		failFatal(t, "Incorrect error value", want, got)
	}

	if got != nil && want != nil {
		if got.Error() != want.Error() {
			failFatal(t, "Incorrect error value", want, got)
		}
	}
}

func compare(t *testing.T, c testCase, got interface{}, err error) {
	checkErr(t, err, c.ExpectedError)

	switch reflect.DeepEqual(got, c.ExpectedResult) {
	case true:
		pass(t, "Test passed", c.ExpectedResult, got)
	case false:
		failFatal(t, "Test failed", c.ExpectedResult, got, c.ExpectedError)
	}
}

func compareSlices(t *testing.T, want, got []int, gotErr, wantErr error) {
	checkErr(t, gotErr, wantErr)

	switch reflect.DeepEqual(got, want) {
	case true:
		pass(t, "Test passed", want, got)
	case false:
		failFatal(t, "Test failed", want, got)
	}
}

func compareFloatSlices(t *testing.T, want, got []float64, gotErr, wantErr error) {
	checkErr(t, gotErr, wantErr)

	switch reflect.DeepEqual(got, want) {
	case true:
		pass(t, "Test passed", want, got)
	case false:
		failFatal(t, "Test failed", want, got)
	}
}

// MARK - Tests

func TestWrontDest(t *testing.T) {
	foo := struct{}{}
	r := newRequest("foo")

	err := Parse(foo, r)

	if err == ErrWrongDestType {
		pass(t, "Test pass", ErrWrongDestType, err)
	} else {
		failFatal(t, "Test pass", ErrWrongDestType, err)
	}
}

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

func TestParseMap(t *testing.T) {
	type testStruct struct {
		Filter Map `qparams:"ops:>,==,<=,<,!=,-like-"`
	}

	table := []testCase{
		{
			URL:            "foobar.com?filter=age>7,gender==0,balance<=1000",
			ExpectedResult: testStruct{Filter: Map{"age >": "7", "gender ==": "0", "balance <=": "1000"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?filter=age>8,gender==1,balance<100",
			ExpectedResult: testStruct{Filter: Map{"age >": "8", "gender ==": "1", "balance <": "100"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?filter=Age>8,Gender==1,Balance<100",
			ExpectedResult: testStruct{Filter: Map{"age >": "8", "gender ==": "1", "balance <": "100"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?filter=,Age>8,Gender==1,Balance<100,",
			ExpectedResult: testStruct{Filter: Map{"age >": "8", "gender ==": "1", "balance <": "100"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?filter=aGe!=9,Gender>0,Lastname-like-Doe",
			ExpectedResult: testStruct{Filter: Map{"age !=": "9", "gender >": "0", "lastname -like-": "Doe"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?filter=",
			ExpectedResult: testStruct{},
			ExpectedError:  nil,
		},
	}

	t.Log("")
	t.Log("Testing map parsing")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)
	}
}

func TestParseMapWithCustomSeparator(t *testing.T) {
	type testStruct struct {
		Filter Map `qparams:"sep:| ops:>,==,<=,<,!=,-like-"`
	}

	table := []testCase{
		{
			URL:            "foobar.com?filter=age>7|gender==0|balance<=1000",
			ExpectedResult: testStruct{Filter: Map{"age >": "7", "gender ==": "0", "balance <=": "1000"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?filter=age>8|gender==1|balance<100",
			ExpectedResult: testStruct{Filter: Map{"age >": "8", "gender ==": "1", "balance <": "100"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?filter=Age>8|Gender==1|Balance<100",
			ExpectedResult: testStruct{Filter: Map{"age >": "8", "gender ==": "1", "balance <": "100"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?filter=|Age>8|Gender==1|Balance<100|",
			ExpectedResult: testStruct{Filter: Map{"age >": "8", "gender ==": "1", "balance <": "100"}},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?filter=aGe!=9|Gender>0|Lastname-like-Doe",
			ExpectedResult: testStruct{Filter: Map{"age !=": "9", "gender >": "0", "lastname -like-": "Doe"}},
			ExpectedError:  nil,
		},
	}

	t.Log("")
	t.Log("Testing map parsing")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)
	}
}

func TestParseInt(t *testing.T) {
	type testStruct struct {
		Limit int
	}

	table := []testCase{
		{
			URL:            "foobar.com?limit=100",
			ExpectedResult: testStruct{Limit: 100},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?limit=0",
			ExpectedResult: testStruct{Limit: 0},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?limit=7",
			ExpectedResult: testStruct{Limit: 7},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?limit=-15",
			ExpectedResult: testStruct{Limit: -15},
			ExpectedError:  nil,
		},
	}

	t.Log("")
	t.Log("Testing integer parsing")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)
	}
}

func TestParseIntErrors(t *testing.T) {
	type testStruct struct {
		Limit int
	}

	table := []testCase{
		{
			URL:            "foobar.com?limit=100a",
			ExpectedResult: testStruct{Limit: 0},
			ExpectedError:  TypeConvErrors{"Field Limit does not contain a valid integer (100a)"},
		},

		{
			URL:            "foobar.com?limit=abc",
			ExpectedResult: testStruct{Limit: 0},
			ExpectedError:  TypeConvErrors{"Field Limit does not contain a valid integer (abc)"},
		},

		{
			URL:            "foobar.com?limit=7,",
			ExpectedResult: testStruct{Limit: 0},
			ExpectedError:  TypeConvErrors{"Field Limit does not contain a valid integer (7,)"},
		},
	}

	t.Log("")
	t.Log("Testing integer parsing errors")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)
	}
}

func TestParseFloat64(t *testing.T) {
	type testStruct struct {
		Ratio float64
	}

	table := []testCase{
		{
			URL:            "foobar.com?ratio=100.876",
			ExpectedResult: testStruct{Ratio: 100.876},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?ratio=0.876",
			ExpectedResult: testStruct{Ratio: 0.876},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?ratio=-100.876",
			ExpectedResult: testStruct{Ratio: -100.876},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?ratio=17.92345",
			ExpectedResult: testStruct{Ratio: 17.92345},
			ExpectedError:  nil,
		},
	}

	t.Log("")
	t.Log("Testing float parsing")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)
	}
}

func TestParseFloatErrors(t *testing.T) {
	type testStruct struct {
		Ratio float64
	}

	table := []testCase{
		{
			URL:            "foobar.com?ratio=100.8a",
			ExpectedResult: testStruct{Ratio: 0},
			ExpectedError:  TypeConvErrors{"Field Ratio does not contain a valid float (100.8a)"},
		},

		{
			URL:            "foobar.com?ratio=a100.8a",
			ExpectedResult: testStruct{Ratio: 0},
			ExpectedError:  TypeConvErrors{"Field Ratio does not contain a valid float (a100.8a)"},
		},

		{
			URL:            "foobar.com?ratio=100_8",
			ExpectedResult: testStruct{Ratio: 0},
			ExpectedError:  TypeConvErrors{"Field Ratio does not contain a valid float (100_8)"},
		},
	}

	t.Log("")
	t.Log("Testing float parsing errors")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)
	}
}

func TestParseString(t *testing.T) {
	type testStruct struct {
		Name string
	}

	table := []testCase{
		{
			URL:            "foobar.com?name=John",
			ExpectedResult: testStruct{Name: "John"},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?name=john",
			ExpectedResult: testStruct{Name: "john"},
			ExpectedError:  nil,
		},

		{
			URL:            "foobar.com?name=john72",
			ExpectedResult: testStruct{Name: "john72"},
			ExpectedError:  nil,
		},
	}

	t.Log("")
	t.Log("Testing string parsing")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)
	}
}

func TestSliceConvert(t *testing.T) {
	type testStruct struct {
		IDs Slice
	}

	table := []testCase{
		{
			URL:                    "foobar.com?ids=1,2,3,7,101",
			ExpectedResult:         testStruct{IDs: Slice{"1", "2", "3", "7", "101"}},
			ExpectedError:          nil,
			ExpectedIntSliceResult: []int{1, 2, 3, 7, 101},
			ExpectedConvErr:        nil,
		},

		{
			URL:                    "foobar.com?ids=,1,2,5,7,101,",
			ExpectedResult:         testStruct{IDs: Slice{"1", "2", "5", "7", "101"}},
			ExpectedError:          nil,
			ExpectedIntSliceResult: []int{1, 2, 5, 7, 101},
			ExpectedConvErr:        nil,
		},

		{
			URL:                    "foobar.com?ids=,1,2,5a,7,101,",
			ExpectedResult:         testStruct{IDs: Slice{"1", "2", "5a", "7", "101"}},
			ExpectedError:          nil,
			ExpectedIntSliceResult: []int{1, 2, 7, 101},
			ExpectedConvErr:        fmt.Errorf("Could not convert member 5a to int"),
		},
	}

	t.Log("")
	t.Log("Testing conversion of qparams slice to string slice")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		Parse(&opts, r)

		got := opts.IDs.Slice()
		want := []string(opts.IDs)

		switch reflect.DeepEqual(got, want) {
		case true:
			pass(t, "Test passed", want, got)
		case false:
			failFatal(t, "Test failed", want, got)
		}

	}
}

func TestIntSliceConvert(t *testing.T) {
	type testStruct struct {
		IDs Slice
	}

	table := []testCase{
		{
			URL:                    "foobar.com?ids=1,2,3,7,101",
			ExpectedResult:         testStruct{IDs: Slice{"1", "2", "3", "7", "101"}},
			ExpectedError:          nil,
			ExpectedIntSliceResult: []int{1, 2, 3, 7, 101},
			ExpectedConvErr:        nil,
		},

		{
			URL:                    "foobar.com?ids=,1,2,5,7,101,",
			ExpectedResult:         testStruct{IDs: Slice{"1", "2", "5", "7", "101"}},
			ExpectedError:          nil,
			ExpectedIntSliceResult: []int{1, 2, 5, 7, 101},
			ExpectedConvErr:        nil,
		},

		{
			URL:                    "foobar.com?ids=,1,2,5a,7,101,",
			ExpectedResult:         testStruct{IDs: Slice{"1", "2", "5a", "7", "101"}},
			ExpectedError:          nil,
			ExpectedIntSliceResult: []int{1, 2, 7, 101},
			ExpectedConvErr:        fmt.Errorf("Could not convert member 5a to int"),
		},
	}

	t.Log("")
	t.Log("Testing conversion of slice to int slice")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)

		newSlice, err := opts.IDs.ToIntSlice()

		compareSlices(t, c.ExpectedIntSliceResult, newSlice, err, c.ExpectedConvErr)
	}
}

func TestFloatSliceConvert(t *testing.T) {
	type testStruct struct {
		IDs Slice
	}

	table := []testCase{
		{
			URL:                      "foobar.com?ids=1,2,3.5,7,101",
			ExpectedResult:           testStruct{IDs: Slice{"1", "2", "3.5", "7", "101"}},
			ExpectedError:            nil,
			ExpectedFloatSliceResult: []float64{1, 2, 3.5, 7, 101},
			ExpectedConvErr:          nil,
		},

		{
			URL:                      "foobar.com?ids=,1,2.345,5,7,101,",
			ExpectedResult:           testStruct{IDs: Slice{"1", "2.345", "5", "7", "101"}},
			ExpectedError:            nil,
			ExpectedFloatSliceResult: []float64{1, 2.345, 5, 7, 101},
			ExpectedConvErr:          nil,
		},

		{
			URL:                      "foobar.com?ids=,1,2.75,5a.7,7,101,",
			ExpectedResult:           testStruct{IDs: Slice{"1", "2.75", "5a.7", "7", "101"}},
			ExpectedError:            nil,
			ExpectedFloatSliceResult: []float64{1, 2.75, 7, 101},
			ExpectedConvErr:          fmt.Errorf("Could not convert member 5a.7 to float"),
		},
	}

	t.Log("")
	t.Log("Testing conversion of slice to float slice")

	for _, c := range table {
		opts := testStruct{}
		r := newRequest(c.URL)
		err := Parse(&opts, r)

		compare(t, c, opts, err)

		newSlice, err := opts.IDs.ToFloatSlice()

		compareFloatSlices(t, c.ExpectedFloatSliceResult, newSlice, err, c.ExpectedConvErr)
	}
}
