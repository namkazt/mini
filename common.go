package mini

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// Sync records fields data
func Sync(from interface{}, to interface{}) interface{} {
	_from := reflect.ValueOf(from)
	_fromType := _from.Type()
	_to := reflect.ValueOf(to)

	for i := 0; i < _from.NumField(); i++ {
		fromName := _fromType.Field(i).Name
		toField := _to.Elem().FieldByName(fromName)
		// move on if toField can't set
		if !toField.IsValid() || !toField.CanSet() {
			continue
		}
		// -----------------------------
		fromKind := _from.Field(i).Kind()
		toKind := toField.Kind()
		// -----------------------------
		// in case of fromKind and toKind is same type
		if fromKind == toKind {
			// if both is pointer then need check if toField is valid and can set or not and from field must not be nil
			if fromKind == reflect.Ptr {
				if !_from.Field(i).IsNil() {
					toField.Set(_from.Field(i))
				}
			} else {
				// else if both is interface{} then
				// TODO: maybe need convert
				toField.Set(_from.Field(i))
			}
		} else {
			// in case of difference type then we need convert kind here
			if fromKind == reflect.Ptr {
				if !_from.Field(i).IsNil() {
					temp := _from.Field(i).Elem().Interface()
					// if from kind is pointer than toKind must be interface{}
					toField.Set(reflect.ValueOf(temp))
				}
			} else {
				temp := _from.Field(i).Interface()
				// if from kind is interface{} then toKind must be pointer
				toField.Set(reflect.ValueOf(temp))
			}
		}
	}
	return to
}

func Jsonb(v string) postgres.Jsonb {
	return postgres.Jsonb{RawMessage: json.RawMessage(v)}
}

func FloorOf(a, b int) int {
	return int(math.Floor(float64(a) / float64(b)))
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MustBeInt(v string) int {
	res, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return res
}
