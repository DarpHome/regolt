package regolt

func P[T any](t T) *T {
	return &t
}

type ArshalNotImplemented struct{}

func (ArshalNotImplemented) Error() string {
	return "Arshaling not implemented"
}

type Marshal func(any) ([]byte, error)
type Unmarshal func([]byte, any) error

type JSONArshaler interface {
	// if you want use default arshal, do `nil, ArshalNotImplemented{}`
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
}

type jsonArshalerImpl struct {
	marshal   Marshal
	unmarshal Unmarshal
}

func (a *jsonArshalerImpl) Marshal(t any) ([]byte, error) {
	if a.marshal != nil {
		return a.marshal(t)
	}
	return nil, ArshalNotImplemented{}
}

func (a *jsonArshalerImpl) Unmarshal(d []byte, t any) error {
	if a.marshal != nil {
		return a.unmarshal(d, t)
	}
	return ArshalNotImplemented{}
}

// both params can be nil
func NewJSONArshaler(marshal Marshal, unmarshal Unmarshal) JSONArshaler {
	return &jsonArshalerImpl{marshal, unmarshal}
}
