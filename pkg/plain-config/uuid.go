package plain_config

import (
	"reflect"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UUID = uuid.UUID

func cast2uuid(data any) (ret UUID, e error) {
	v := reflect.Indirect(reflect.ValueOf(data))
	switch vt := v.Type(); vt.Kind() {
	case reflect.String, reflect.Slice:
		tyString := reflect.TypeOf((*string)(nil)).Elem()
		d := v.Convert(tyString).Interface().(string)
		return uuid.Parse(d)
	case reflect.Array:
		tyBase := reflect.TypeOf((*UUID)(nil)).Elem()
		if vt.ConvertibleTo(tyBase) {
			reflect.ValueOf(&ret).Elem().Set(
				v.Convert(tyBase),
			)
			return ret, nil
		}
	}
	return ret, errors.Errorf("unable covert '%v' into UUID", data)
}
