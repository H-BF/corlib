package jobs

import (
	"fmt"
)

type errMarshal struct {
	error
}

func (err errMarshal) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", err.String())), nil
}

func (err errMarshal) Error() string {
	if err.error == nil {
		return "null"
	}
	return err.error.Error()
}

func (err errMarshal) String() string {
	if i, ok := err.error.(fmt.Stringer); ok && err.error != nil {
		return i.String()
	}
	return err.Error()
}
