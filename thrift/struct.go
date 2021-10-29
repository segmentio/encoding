package thrift

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type flags int16

const (
	noflags flags = 0
	enum    flags = 1 << 0
	union   flags = 1 << 1
)

func (f flags) have(x flags) bool {
	return (f & x) == x
}

func (f flags) with(x flags) flags {
	return f | x
}

type structField struct {
	typ   reflect.Type
	index []int
	id    int16
	flags flags
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
		flags := flags(0)

		for _, opt := range tags[1:] {
			switch opt {
			case "enum":
				flags = flags.with(enum)
			case "union":
				flags = flags.with(union)
			}
		}

		if flags.have(enum) {
			switch f.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			default:
				panic(fmt.Errorf("thrift enum tag found on a field which is not an integer type `%s`", tag))
			}
		}

		if flags.have(union) {
			if f.Type.Kind() != reflect.Struct {
				panic(fmt.Errorf("thrift union tag found on a field which does not have a struct type `%s`", tag))
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
				flags: flags,
			})
		}
	}
}
