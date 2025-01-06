package qs_test

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"

	"github.com/dmji/qs"
)

func ExampleMarshal() {
	type Query struct {
		Search   string
		Page     int
		PageSize int
	}

	queryString, err := qs.Marshal(&Query{
		Search:   "my search",
		Page:     2,
		PageSize: 50,
	})

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(queryString)
	}
	// Output:
	// page=2&page_size=50&search=my+search
}

func ExampleMarshalValues() {
	type Query struct {
		Search   string
		Page     int
		PageSize int
	}

	// values is a url.Values which is a map[string][]string
	values, err := qs.MarshalValues(&Query{
		Search:   "my search",
		Page:     2,
		PageSize: 50,
	})

	if err != nil {
		fmt.Println(err)
	} else {
		// printing the values map after sorting its keys
		var keys []string
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Printf("%v: %v\n", key, values[key])
		}
	}
	// Output:
	// page: [2]
	// page_size: [50]
	// search: [my search]
}

func ExampleCheckMarshal() {
	m := map[string]int{}

	fmt.Println(qs.CheckMarshal(0))
	fmt.Println(qs.CheckMarshal(m))
	fmt.Println(qs.CheckMarshal(&m))
	// Output:
	// unhandled type: int
	// <nil>
	// <nil>
}

func ExampleCheckMarshalType() {
	intType := reflect.TypeOf(0)
	mapType := reflect.TypeOf((map[string]int)(nil))

	fmt.Println(qs.CheckMarshalType(intType))
	fmt.Println(qs.CheckMarshalType(mapType))
	fmt.Println(qs.CheckMarshalType(reflect.PointerTo(mapType)))
	// Output:
	// unhandled type: int
	// <nil>
	// <nil>
}

func ExampleCheckUnmarshal() {
	m := map[string]int{}

	fmt.Println(qs.CheckUnmarshal(0))
	fmt.Println(qs.CheckUnmarshal(m))
	fmt.Println(qs.CheckUnmarshal(&m))
	// Output:
	// expected a pointer, got int
	// expected a pointer, got map[string]int
	// <nil>
}

func ExampleCheckUnmarshalType() {
	intType := reflect.TypeOf(0)
	mapType := reflect.TypeOf((map[string]int)(nil))

	fmt.Println(qs.CheckUnmarshalType(intType))
	fmt.Println(qs.CheckUnmarshalType(mapType))
	fmt.Println(qs.CheckUnmarshalType(reflect.PointerTo(mapType)))
	// Output:
	// expected a pointer, got int
	// expected a pointer, got map[string]int
	// <nil>
}

func ExampleUnmarshal() {
	type Query struct {
		Search   string
		Page     int
		PageSize int
	}

	var q Query
	err := qs.Unmarshal(&q, "page=2&page_size=50&search=my+search")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(q)
	}
	// Output:
	// {my search 2 50}
}

func ExampleUnmarshalValues() {
	type Query struct {
		Search   string
		Page     int
		PageSize int
	}

	var q Query
	err := qs.UnmarshalValues(&q, url.Values{
		"search":    {"my search"},
		"page":      {"2"},
		"page_size": {"50"},
	})

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(q)
	}
	// Output:
	// {my search 2 50}
}
