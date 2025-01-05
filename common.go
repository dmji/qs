package qs

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
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
type NameTransformFunc func(string) string

var stringType = reflect.TypeOf("")
var timeType = reflect.TypeOf(time.Time{})
var urlType = reflect.TypeOf(url.URL{})

type parsedTag struct {
	Name              string
	MarshalPresence   MarshalPresence
	UnmarshalPresence UnmarshalPresence
}

func getStructFieldInfo(field reflect.StructField, nt NameTransformFunc, defaultMarshalPresence MarshalPresence, defaultUnmarshalPresence UnmarshalPresence) (skip bool, tag parsedTag, err error) {
	// Skipping unexported fields.
	if field.PkgPath != "" && !field.Anonymous {
		skip = true
		return
	}

	tag, err = parseFieldTag(field.Tag, defaultMarshalPresence, defaultUnmarshalPresence)
	if err != nil {
		err = fmt.Errorf("invalid tag: %q :: %v", field.Tag, err)
		return
	}

	// Skipping this field if the tag specifies "-" as field name.
	if tag.Name == "-" {
		skip = true
		return
	}

	if tag.Name == "" {
		tag.Name = nt(field.Name)
	}

	return
}

func parseFieldTag(tagStr reflect.StructTag, defaultMarshalPresence MarshalPresence, defaultUnmarshalPresence UnmarshalPresence) (tag parsedTag, err error) {
	v := tagStr.Get(tagKey)
	arr := strings.Split(v, ",")
	tag.Name = arr[0]

	setMarshalPresence := func(v MarshalPresence) {
		if tag.MarshalPresence != MarshalPresenceMPUnspecified {
			err = fmt.Errorf("only one MarshalPresence option is allwed - you've specified at least two: %v, %v", tag.MarshalPresence, v)
		}
		tag.MarshalPresence = v
	}

	setUnmarshalPresence := func(v UnmarshalPresence) {
		if tag.UnmarshalPresence != UnmarshalPresenceUPUnspecified {
			err = fmt.Errorf("only one UnmarshalPresence option is allwed - you've specified at least two: %v, %v", tag.UnmarshalPresence, v)
		}
		tag.UnmarshalPresence = v
	}

	for _, option := range arr[1:] {
		switch option {
		case "nil":
			setUnmarshalPresence(UnmarshalPresenceNil)
		case "opt":
			setUnmarshalPresence(UnmarshalPresenceOpt)
		case "req":
			setUnmarshalPresence(UnmarshalPresenceReq)
		case "keepempty":
			setMarshalPresence(MarshalPresenceKeepEmpty)
		case "omitempty":
			setMarshalPresence(MarshalPresenceOmitEmpty)
		case "":
			err = errors.New("tag string contains a surplus comma")
		default:
			err = fmt.Errorf("invalid option in field tag: %q", option)
		}
		if err != nil {
			return
		}
	}

	if tag.MarshalPresence == MarshalPresenceMPUnspecified {
		tag.MarshalPresence = defaultMarshalPresence
	}
	if tag.UnmarshalPresence == UnmarshalPresenceUPUnspecified {
		tag.UnmarshalPresence = defaultUnmarshalPresence
	}

	return
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
