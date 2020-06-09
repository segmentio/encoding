package proto

type Message interface {
	Size() int

	Marshal([]byte) error

	Unmarshal([]byte) error
}

type RawMessage []byte

func (m RawMessage) Size() int { return len(m) }

func (m RawMessage) Marshal(b []byte) error {
	copy(b, m)
	return nil
}

func (m *RawMessage) Unmarshal(b []byte) error {
	*m = make([]byte, len(b))
	copy(*m, b)
	return nil
}
