package thrift

type Message struct {
	Type  MessageType
	Name  string
	SeqID int32
}

type MessageType int8

const (
	Call MessageType = iota
	Reply
	Exception
	Oneway
)

type Field struct {
	ID   int16
	Type Type
}

type Type int8

const (
	TRUE Type = iota + 1
	FALSE
	I8
	I16
	I32
	I64
	DOUBLE
	BINARY
	LIST
	SET
	MAP
	STRUCT
	BOOL = FALSE
)

type List struct {
	Size int32
	Type Type
}

type Map struct {
	Size  int32
	Key   Type
	Value Type
}
