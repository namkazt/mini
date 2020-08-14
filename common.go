package mini

import (
	"encoding/json"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"time"

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

func IsMobile(mobile string) bool {
	l := len(mobile)
	if l == 0 {
		return false
	}
	isInt := func(in string) bool {
		_, err := strconv.ParseInt(in, 10, 64)
		return err == nil
	}
	if mobile[0] == '+' && l == 13 {
		countryCode := mobile[3:]
		if countryCode != "88" {
			return false
		}
		mobile = "0" + mobile[3:]
		if !isInt(mobile) {
			return false
		}
		return true
	} else if mobile[0] == '0' && l == 11 {
		if !isInt(mobile) {
			return false
		}
		return true
	}
	return false
}

const (
	TL_API_letterBytes   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	TL_API_numberBytes   = "0123456789"
	TL_API_letterIdxBits = 6                           // 6 bits to represent a letter index
	TL_API_letterIdxMask = 1<<TL_API_letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	TL_API_letterIdxMax  = 63 / TL_API_letterIdxBits   // # of letter indices fitting in 63 bits
)

func random(length int, source string) string {
	//===============================================================================
	// We using this method:
	// source: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
	// NOTE:
	// If using it on multi-threads or goroutine better pass each thread/goroutine 1 Random source
	// or for simple just add Mutex lock here to lock source when using it.
	//===============================================================================
	var TL_API_SRC_RANDOM = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, length)
	for i, cache, remain := length-1, TL_API_SRC_RANDOM.Int63(), TL_API_letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = TL_API_SRC_RANDOM.Int63(), TL_API_letterIdxMax
		}
		if idx := int(cache & TL_API_letterIdxMask); idx < len(source) {
			b[i] = source[idx]
			i--
		}
		cache >>= TL_API_letterIdxBits
		remain--
	}
	return string(b)
}

func RandomString(length int) string {
	return random(length, TL_API_letterBytes)
}

func RandomNumber(length int) string {
	return random(length, TL_API_numberBytes)
}
