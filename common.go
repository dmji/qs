package qs

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"
)

const tagKey = "qs"

// A NameTransformFunc is used to derive the query string keys from the field
// names of the struct.
// NameTransformFunc is the type of the DefaultNameTransform,
// MarshalOptions.NameTransformer and UnmarshalOptions.NameTransformer variables.
type (
	NameTransformFunc func(string) string
	SliceToStringFunc func([]string) (string, error)
)

var (
	stringType = reflect.TypeOf("")
	timeType   = reflect.TypeOf(time.Time{})
	urlType    = reflect.TypeOf(url.URL{})
)

type ParsedTagInfo struct {
	Name                          string
	MarshalPresence               MarshalPresence
	UnmarshalPresence             UnmarshalPresence
	UnmarshalSliceValues          UnmarshalSliceValues
	UnmarshalSliceUnexpectedValue UnmarshalSliceUnexpectedValue
}

func getStructFieldInfo(field reflect.StructField, nt NameTransformFunc, defaultMarshalPresence MarshalPresence, defaultUnmarshalPresence UnmarshalPresence) (*ParsedTagInfo, error) {
	// Skipping unexported fields.
	if field.PkgPath != "" && !field.Anonymous {
		return nil, nil
	}

	tag, err := parseFieldTag(field.Tag, defaultMarshalPresence, defaultUnmarshalPresence)
	if err != nil {
		err = fmt.Errorf("invalid tag: %q :: %v", field.Tag, err)
		return nil, err
	}

	// Skipping this field if the tag specifies "-" as field name.
	if tag.Name == "-" {
		return nil, nil
	}

	if tag.Name == "" {
		tag.Name = nt(field.Name)
	}

	return tag, nil
}

const fmtOptionNotUniqueError = "only one %s option is allwed - you've specified at least two: %v, %v"

func parseFieldTag(tagStr reflect.StructTag, defaultMarshalPresence MarshalPresence, defaultUnmarshalPresence UnmarshalPresence) (*ParsedTagInfo, error) {
	v := tagStr.Get(tagKey)
	nameAndOptions := strings.Split(v, ",")
	tag := &ParsedTagInfo{
		Name:                          nameAndOptions[0],
		MarshalPresence:               MarshalPresenceMPUnspecified,
		UnmarshalPresence:             UnmarshalPresenceUPUnspecified,
		UnmarshalSliceValues:          UnmarshalSliceValuesUPUnspecified,
		UnmarshalSliceUnexpectedValue: UnmarshalSliceUnexpectedValueUPUnspecified,
	}

	options := nameAndOptions[1:]
	if slices.IndexFunc(options, func(i string) bool { return len(i) == 0 }) != -1 {
		return nil, errors.New("tag string contains a surplus comma")
	}

	for _, option := range options {
		bErr := true

		// UnmarshalPresence
		if value, err := UnmarshalPresenceFromString(option); err == nil {
			if tag.UnmarshalPresence != UnmarshalPresenceUPUnspecified {
				return nil, fmt.Errorf(fmtOptionNotUniqueError, "UnmarshalPresence", tag.UnmarshalPresence, v)
			}
			tag.UnmarshalPresence = value
			bErr = false
		}

		// UnmarshalSliceValues
		if value, err := UnmarshalSliceValuesFromString(option); err == nil {
			if tag.UnmarshalSliceValues != UnmarshalSliceValuesUPUnspecified {
				return nil, fmt.Errorf(fmtOptionNotUniqueError, "UnmarshalSliceValues", tag.UnmarshalSliceValues, v)
			}
			tag.UnmarshalSliceValues = value
			bErr = false
		}

		// UnmarshalSliceValues
		if value, err := UnmarshalSliceUnexpectedValueFromString(option); err == nil {
			if tag.UnmarshalSliceUnexpectedValue != UnmarshalSliceUnexpectedValueUPUnspecified {
				return nil, fmt.Errorf(fmtOptionNotUniqueError, "UnmarshalSliceUnexpectedValue", tag.UnmarshalSliceUnexpectedValue, v)
			}
			tag.UnmarshalSliceUnexpectedValue = value
			bErr = false
		}

		// MarshalPresence
		if value, err := MarshalPresenceFromString(option); err == nil {
			if tag.MarshalPresence != MarshalPresenceMPUnspecified {
				return nil, fmt.Errorf(fmtOptionNotUniqueError, "MarshalPresence", tag.MarshalPresence, v)
			}
			tag.MarshalPresence = value
			bErr = false
		}

		// Error specified option name is invalid
		if bErr {
			return nil, fmt.Errorf("invalid option in field tag: %q", option)
		}
	}

	if tag.MarshalPresence == MarshalPresenceMPUnspecified {
		tag.MarshalPresence = defaultMarshalPresence
	}
	if tag.UnmarshalPresence == UnmarshalPresenceUPUnspecified {
		tag.UnmarshalPresence = defaultUnmarshalPresence
	}

	return tag, nil
}

// snakeCase converts CamelCase names to snake_case with lowercase letters and
// underscores. Names already in snake_case are left untouched.
func snakeCase(s string) string {
	in := []rune(s)
	isLower := func(idx int) bool {
		return idx >= 0 && idx < len(in) && unicode.IsLower(in[idx])
	}

	out := make([]rune, 0, len(in)+len(in)/2)
	for i, r := range in {
		if unicode.IsUpper(r) {
			r = unicode.ToLower(r)
			if i > 0 && in[i-1] != '_' && (isLower(i-1) || isLower(i+1)) {
				out = append(out, '_')
			}
		}
		out = append(out, r)
	}

	return string(out)
}

func cacher[TRes any, TOpt any](wrapped func(t reflect.Type, opts *TOpt) (TRes, error), cache *sync.Map, t reflect.Type, opts *TOpt) (TRes, error) {
	var (
		m   TRes
		err error
	)
	if item, ok := cache.Load(t); ok {
		if m, ok = item.(TRes); ok {
			return m, nil
		}
		return m, item.(error)
	}

	m, err = wrapped(t, opts)
	if err != nil {
		cache.Store(t, err)
	} else {
		cache.Store(t, m)
	}
	return m, err
}
