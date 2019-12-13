// +build go1.14

package json

import (
	"reflect"
	"unsafe"
)

func extendSlice(t reflect.Type, s *slice, n int) slice {
	arrayType := reflect.ArrayOf(t.Elem(), n)
	arrayData := reflect.New(arrayType)
	reflect.Copy(arrayData.Elem(), reflect.NewAt(t, unsafe.Pointer(s)))
	return slice{
		data: unsafe.Pointer(arrayData.Pointer()),
		len:  s.len,
		cap:  n,
	}
}
