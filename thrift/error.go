package thrift

import (
	"fmt"
	"strings"
)

type MissingField struct {
	Field Field
}

func (e *MissingField) Error() string {
	return fmt.Sprintf("missing required field: %s", e.Field)
}

type TypeMismatch struct {
	Expect Type
	Found  Type
	item   string
}

func (e *TypeMismatch) Error() string {
	return fmt.Sprintf("%s type mismatch: expected %s but found %s", e.item, e.Expect, e.Found)
}

type decodeError struct {
	base error
	path []error
}

func (e *decodeError) Error() string {
	s := strings.Builder{}
	s.Grow(256)
	s.WriteString("decoding thrift payload: ")

	if len(e.path) != 0 {
		n := len(e.path) - 1
		for i := n; i >= 0; i-- {
			if i < n {
				s.WriteString(" → ")
			}
			s.WriteString(e.path[i].Error())
		}
		s.WriteString(": ")
	}

	s.WriteString(e.base.Error())
	return s.String()
}

func (e *decodeError) Unwrap() error { return e.base }

func with(base, elem error) error {
	e, _ := base.(*decodeError)
	if e == nil {
		e = &decodeError{base: base}
	}
	e.path = append(e.path, elem)
	return e
}

type decodeErrorField struct {
	field Field
}

func (d *decodeErrorField) Error() string {
	return d.field.String()
}

type decodeErrorList struct {
	list  List
	index int
}

func (d *decodeErrorList) Error() string {
	return fmt.Sprintf("%d/%d:%s", d.index, d.list.Size, d.list)
}

type decodeErrorSet struct {
	set   Set
	index int
}

func (d *decodeErrorSet) Error() string {
	return fmt.Sprintf("%d/%d:%s", d.index, d.set.Size, d.set)
}

type decodeErrorMap struct {
	_map  Map
	index int
}

func (d *decodeErrorMap) Error() string {
	return fmt.Sprintf("%d/%d:%s", d.index, d._map.Size, d._map)
}
