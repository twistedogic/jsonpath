package jsonpath

import (
	"reflect"
	"strings"

	"github.com/fatih/structs"
	json "github.com/json-iterator/go"
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

func getTaggedField(i interface{}, parentOmit bool) []Field {
	fields := structs.New(i).Fields()
	out := make([]Field, 0, len(fields))
	for _, f := range fields {
		tag := f.Tag(DefaultTagName)
		if tag != "" {
			tokens := strings.Split(tag, ",")
			path, other := tokens[0], tokens[1:]
			omit := false
			for _, o := range other {
				if strings.TrimSpace(o) == OmitEmpty {
					omit = true
					break
				}
			}
			out = append(out, Field{
				Name:      f.Name(),
				Value:     f.Value(),
				Pattern:   path,
				OmitEmpty: omit || parentOmit,
			})
		}
	}
	return out
}

func IsValue(in interface{}) bool {
	if in == nil {
		return false
	}
	inKind := reflect.TypeOf(in).Kind()
	switch {
	case inKind == reflect.Slice:
		return false
	case inKind == reflect.Struct:
		return false
	}
	return true
}

func parseJsonpath(in interface{}, out interface{}, omit bool) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	fields := getTaggedField(out, omit)
	for _, f := range fields {
		value, err := Lookup(f.Pattern, in)
		switch {
		case err != nil && !f.OmitEmpty:
			return obj, err
		case value == nil:
			continue
		case err == nil:
			if structs.IsStruct(f.Value) {
				if nested, err := parseJsonpath(value, f.Value, f.OmitEmpty); err != nil {
					return obj, err
				} else {
					obj[f.Name] = nested
				}
			} else {
				if !IsValue(value) && f.OmitEmpty {
					obj[f.Name] = f.Value
				} else {
					obj[f.Name] = value
				}
			}
		}
	}
	return obj, nil
}

func ParseJsonpath(in interface{}, out interface{}) error {
	obj, err := parseJsonpath(in, out, false)
	if err != nil {
		return err
	}
	return mapstructure.WeakDecode(obj, out)
}

func Unmarshal(b []byte, i interface{}) error {
	switch v := i.(type) {
	case Unmarshaler:
		return v.UnmarshalJSONPath(b)
	}
	var in interface{}
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}
	return ParseJsonpath(in, i)
}
