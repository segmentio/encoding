package thrift

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type structField struct {
	typ   reflect.Type
	index []int
	id    int16
	enum  bool
}

func forEachStructField(t reflect.Type, index []int, do func(structField)) {
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

		tag := f.Tag.Get("thrift")
		if tag == "" {
			continue
		}
		tags := strings.Split(tag, ",")
		enum := false

		for _, opt := range tags[1:] {
			switch opt {
			case "enum":
				enum = true
			}
		}

		if id, err := strconv.ParseInt(tags[0], 10, 16); err != nil {
			panic(fmt.Errorf("invalid thrift field id found in struct tag `%s`: %w", tag, err))
		} else if id <= 0 {
			panic(fmt.Errorf("invalid thrift field id found in struct tag `%s`: %d <= 0", tag, id))
		} else {
			do(structField{
				typ:   f.Type,
				index: fieldIndex,
				id:    int16(id),
				enum:  enum,
			})
		}
	}
}
