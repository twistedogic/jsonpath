package jsonpath

import (
	"encoding/json"
	"strings"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"github.com/oliveagle/jsonpath"
)

const (
	DefaultTagName = "jsonpath"
	OmitEmpty      = "omitempty"
)

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

type Field struct {
	Name      string
	Value     interface{}
	Pattern   string
	OmitEmpty bool
}

func getTaggedField(i interface{}) []Field {
	fields := structs.New(i).Fields()
	out := make([]Field, 0, len(fields))
	for _, f := range fields {
		tag := f.Tag(DefaultTagName)
		if tag != "" {
			tokens := strings.Split(tag, ",")
			path, other := tokens[0], tokens[1:]
			omit := false
			if len(other) > 0 {
				omit = other[0] == OmitEmpty
			}
			out = append(out, Field{
				Name:      f.Name(),
				Value:     f.Value(),
				Pattern:   path,
				OmitEmpty: omit,
			})
		}
	}
	return out
}

func parseJsonpath(in interface{}, out interface{}) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	fields := getTaggedField(out)
	for _, f := range fields {
		value, err := Lookup(f.Pattern, in)
		switch {
		case err != nil && !f.OmitEmpty:
			return obj, err
		case err == nil:
			if structs.IsStruct(f.Value) {
				nested, err := parseJsonpath(value, f.Value)
				if err != nil {
					return obj, err
				}
				obj[f.Name] = nested
			} else {
				obj[f.Name] = value
			}
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
