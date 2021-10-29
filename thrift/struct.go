package thrift

import (
	"fmt"
	"reflect"
	"strconv"
)

func forEachStructField(t reflect.Type, index []int, do func(reflect.Type, int16, []int)) {
	for i, n := 0, t.NumField(); i < n; i++ {
		f := t.Field(i)

		if !f.IsExported() {
			continue
		}

		fieldIndex := append(index, i)
		fieldIndex = fieldIndex[:len(fieldIndex):len(fieldIndex)]

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			forEachStructField(f.Type, fieldIndex, do)
			continue
		}

		if tag := f.Tag.Get("thrift"); tag == "" {
			continue
		} else if id, err := strconv.ParseInt(tag, 10, 16); err != nil {
			panic(fmt.Errorf("invalid thrift field id found in struct tag: %w", err))
		} else if id <= 0 {
			panic(fmt.Errorf("invalid thrift field id found in struct tag: %d <= 0", id))
		} else {
			do(f.Type, int16(id), fieldIndex)
		}
	}
}
