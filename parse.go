package jsonpath

import (
	"encoding/json"
	"strings"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"github.com/oliveagle/jsonpath"
)

const DefaultTagName = "jsonpath"

type Unmarshaler interface {
	UnmarshalJSONPath([]byte) error
}

func Lookup(pattern string, i interface{}) (interface{}, error) {
	jpath, err := jsonpath.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return jpath.Lookup(i)
}

func parseJsonpath(in interface{}, out interface{}) (map[string]interface{}, error) {
	fields := structs.New(out).Fields()
	obj := make(map[string]interface{})
	for _, f := range fields {
		tag := f.Tag(DefaultTagName)
		fieldValue := f.Value()
		fieldName := f.Name()
		tokens := strings.Split(tag, ",")
		pattern, _ := tokens[0], tokens[1:]
		value, err := Lookup(pattern, in)
		if err != nil {
			return obj, err
		}
		if structs.IsStruct(fieldValue) {
			nested, err := parseJsonpath(value, fieldValue)
			if err != nil {
				return obj, err
			}
			obj[fieldName] = nested
		} else {
			obj[fieldName] = value
		}
	}
	return obj, nil
}

func ParseJsonpath(in interface{}, out interface{}) error {
	obj, err := parseJsonpath(in, out)
	if err != nil {
		return err
	}
	return mapstructure.WeakDecode(obj, out)
}

func Unmarshal(b []byte, i interface{}) error {
	switch v := i.(type) {
	case Unmarshaler:
		return v.UnmarshalJSONPath(b)
	case json.Unmarshaler:
		return v.UnmarshalJSON(b)
	}
	var in interface{}
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}
	return ParseJsonpath(in, i)
}
