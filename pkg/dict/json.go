package dict

import (
	"encoding/json"
)

type (
	dictJsonHelper[Tk any, Tv any] struct {
		d Dict[Tk, Tv]
	}
	setJsonHelper[T any] struct {
		s Set[T]
	}
)

// MarshalJSON impl json.Marshaler
func (dict dictJsonHelper[Tk, Tv]) MarshalJSON() ([]byte, error) {
	return json.Marshal(dict.d.Items())
}

// UnmarshalJSON impl json.Unmarshaler
func (dict *dictJsonHelper[Tk, Tv]) UnmarshalJSON(b []byte) error {
	var items Items[Tk, Tv]
	if err := json.Unmarshal(b, &items); err != nil {
		return err
	}
	for i := range items {
		it := items[i]
		dict.d.Put(it.K, it.V)
	}
	return nil
}

// MarshalJSON impl json.Marshaler
func (s setJsonHelper[Tk]) MarshalJSON() ([]byte, error) {
	vals := s.s.Values()
	return json.Marshal(vals)
}

// UnmarshalJSON impl json.Unmarshaler
func (s *setJsonHelper[T]) UnmarshalJSON(b []byte) error {
	var items []T
	if err := json.Unmarshal(b, &items); err != nil {
		return err
	}
	s.s.PutMany(items...)
	return nil
}

// MarshalJSON impl json.Marshaler
func (dict RBDict[Tk, Tv]) MarshalJSON() ([]byte, error) {
	h := dictJsonHelper[Tk, Tv]{d: &dict}
	return h.MarshalJSON()
}

// MarshalJSON impl json.Marshaler
func (dict HDict[Tk, Tv]) MarshalJSON() ([]byte, error) {
	h := dictJsonHelper[Tk, Tv]{d: &dict}
	return h.MarshalJSON()
}

// MarshalJSON impl json.Marshaler
func (s RBSet[T]) MarshalJSON() ([]byte, error) {
	h := setJsonHelper[T]{s: &s}
	return h.MarshalJSON()
}

// MarshalJSON impl json.Marshaler
func (s HSet[T]) MarshalJSON() ([]byte, error) {
	h := setJsonHelper[T]{s: &s}
	return h.MarshalJSON()
}

// UnmarshalJSON impl json.Unmarshaler
func (dict *RBDict[Tk, Tv]) UnmarshalJSON(b []byte) error {
	h := dictJsonHelper[Tk, Tv]{d: dict}
	return h.UnmarshalJSON(b)
}

// UnmarshalJSON impl json.Unmarshaler
func (dict *HDict[Tk, Tv]) UnmarshalJSON(b []byte) error {
	h := dictJsonHelper[Tk, Tv]{d: dict}
	return h.UnmarshalJSON(b)
}

// UnmarshalJSON impl json.Unmarshaler
func (s *HSet[T]) UnmarshalJSON(b []byte) error {
	h := setJsonHelper[T]{s: s}
	return h.UnmarshalJSON(b)
}

// UnmarshalJSON impl json.Unmarshaler
func (s *RBSet[T]) UnmarshalJSON(b []byte) error {
	h := setJsonHelper[T]{s: s}
	return h.UnmarshalJSON(b)
}
