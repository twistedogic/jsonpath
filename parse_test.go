package jsonpath

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Result struct {
	ID    int    `jsonpath:"$.data._id"`
	Value string `jsonpath:"$.data.value"`
}

func TestParseJsonpath(t *testing.T) {
	input := []byte(`{"test":1, "data":{"_id":123123, "value": "something"}}`)
	var in interface{}
	if err := json.Unmarshal(input, &in); err != nil {
		t.Fatal(err)
	}
	expect := Result{Value: "something", ID: 123123}
	var out Result
	if err := ParseJsonpath(in, &out); err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(expect, out) {
		t.Fatalf("want %#v got %#v", expect, out)
	}
}
