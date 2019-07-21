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

func getTaggedField(i interface{}) []*structs.Field {
	fields := structs.New(i).Fields()
	out := make([]*structs.Field, 0, len(fields))
	for _, f := range fields {
		if f.Tag(DefaultTagName) != "" {
			out = append(out, f)
		}
	}
	return out
}

func parseJsonpath(in interface{}, out interface{}) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	fields := getTaggedField(out)
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
