package jsonpath

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Nested struct {
	Name  string `jsonpath:"$.name"`
	Value int    `jsonpath:"$.value, omitempty"`
}

type Result struct {
	ID     int    `jsonpath:"$.data._id"`
	Value  string `jsonpath:"$.data.value"`
	Other  int
	Omit   string `jsonpath:"$.notexistpath,omitempty"`
	Nested Nested `jsonpath:"$.nested, omitempty"`
}

func TestParseJsonpath(t *testing.T) {
	cases := map[string]struct {
		input    []byte
		want     Result
		hasError bool
	}{
		"base": {
			input: []byte(`{
				"test": 1, 
				"data": {
					"_id": 123123, 
					"value": "something"
				}
			}`),
			want: Result{
				Value: "something",
				ID:    123123,
			},
			hasError: false,
		},
		"nested": {
			input: []byte(`{
				"test": 1, 
				"data": {
					"_id": 123123, 
					"value": "something"
				},
				"nested":{
					"name": "name"
				}
			}`),
			want: Result{
				Value: "something",
				ID:    123123,
				Nested: Nested{
					Name: "name",
				},
			},
			hasError: false,
		},
		"extra": {
			input: []byte(`{
				"test": 1, 
				"data": {
					"_id": 123123, 
					"value": "something",
					"extra": "ok"
				},
				"nested":{
					"name": "name"
				}
			}`),
			want: Result{
				Value: "something",
				ID:    123123,
				Nested: Nested{
					Name: "name",
				},
			},
			hasError: false,
		},
		"malform": {
			input: []byte(`{
				"test": 1, 
				"data": {
					"_id": 123123, 
					"value": "something"
				},
				"nested":[]
			}`),
			want: Result{
				Value: "something",
				ID:    123123,
			},
			hasError: false,
		},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			var in interface{}
			if err := json.Unmarshal(tc.input, &in); err != nil {
				t.Fatal(err)
			}
			var got Result
			if err := ParseJsonpath(in, &got); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
