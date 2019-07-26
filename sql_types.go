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
	v.Bool = value
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

// NullInt64
type NullInt64 struct {
	sql.NullInt64
}

// NewNullInt64
func NewNullInt64(value int64) NullInt64 {
	v := NullInt64{}
	v.Int64 = value
	v.Valid = true
	return v
}

// MarshalJSON
func (v *NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	} else {
		return json.Marshal(nil)
	}
}

// UnmarshalJSON
func (v *NullInt64) UnmarshalJSON(data []byte) error {
	var s *int64
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		v.Valid = true
		v.Int64 = *s
	} else {
		v.Valid = false
	}
	return nil
}
