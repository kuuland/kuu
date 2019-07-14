package kuu

import (
	"database/sql"
	"encoding/json"
)

// NullBool
type NullBool struct {
	sql.NullBool
}

// NewNullBool
func NewNullBool(value bool) NullBool {
	v := NullBool{}
	v.Bool = true
	v.Valid = true
	return v
}

// MarshalJSON
func (v *NullBool) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Bool)
	} else {
		return json.Marshal(nil)
	}
}

// UnmarshalJSON
func (v *NullBool) UnmarshalJSON(data []byte) error {
	var s *bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		v.Valid = true
		v.Bool = *s
	} else {
		v.Valid = false
	}
	return nil
}
