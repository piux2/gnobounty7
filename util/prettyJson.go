package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func PrettyJson(a interface{}) {

	v, err := json.MarshalIndent(a, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(v))
}

// Json Line is the format can be imported directly in  BigQuery or PostgreSQL with simple hacks.
func JsonLine(a interface{}) {

	value := reflect.ValueOf(a)
	k := value.Type().Kind()
	if k == reflect.Ptr {
		value = value.Elem()
		k = value.Type().Kind()
	}

	if k == reflect.Array || k == reflect.Slice {
		for i := 0; i < value.Len(); i++ {

			v, err := json.Marshal(value.Index(i).Interface())
			if err != nil {
				panic(err)
			}
			fmt.Printf("%v\n", string(v))

		}
	}

}

// if s contains a keyword in the slices return true and matached keyword

func StringContains(s string, keywords []string) (string, bool) {

	for _, v := range keywords {

		if strings.HasSuffix(s, v) == true {

			return v, true
		}

	}
	return "", false

}
