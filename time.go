package mini

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ======================================================
// Custom Time struct support JSON and GORM
// ======================================================
type Datetime struct {
	time.Time
}

// Json support
// ---------------------------------------------
func (t *Datetime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		t.Time = time.Time{}
		return
	}
	t.Time, err = time.Parse(time.RFC3339, s)
	return
}

func (t *Datetime) MarshalJSON() ([]byte, error) {
	if t.Time.UnixNano() == (time.Time{}).UnixNano() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", t.Time.Format(time.RFC3339))), nil
}

// Gorm support
// ---------------------------------------------
func (t *Datetime) Scan(src interface{}) (err error) {
	var tt Datetime
	switch src.(type) {
	case time.Time:
		tt.Time = src.(time.Time)
	default:
		return errors.New("Incompatible type for Skills")
	}
	*t = tt
	return
}

func (t Datetime) Value() (driver.Value, error) {
	if !t.IsSet() {
		return "null", nil
	}
	return t.Time, nil
}

func (t *Datetime) IsSet() bool {
	return t.UnixNano() != (time.Time{}).UnixNano()
}
