package kuu

//
//import (
//	"encoding/json"
//	"fmt"
//	"reflect"
//)
//
//// ExAttrs
//type ExAttrs struct {
//	raw   string
//	cache map[string]interface{}
//}
//
//// UnmarshalJSON implements json.Unmarshaler.
//func (ex *ExAttrs) UnmarshalJSON(data []byte) error {
//	var err error
//	var v interface{}
//	if err = json.Unmarshal(data, &v); err != nil {
//		return err
//	}
//	switch x := v.(type) {
//	case string:
//		s.String = x
//	case map[string]interface{}:
//		err = json.Unmarshal(data, &s.NullString)
//	case nil:
//		s.Valid = false
//		return nil
//	default:
//		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.String", reflect.TypeOf(v).Name())
//	}
//	s.Valid = err == nil
//	return err
//}
//
//// UnmarshalText implements encoding.TextUnmarshaler.
//func (ex *ExAttrs) UnmarshalText(text []byte) error {
//	s.String = string(text)
//	s.Valid = s.String != ""
//	return nil
//}
//
//// MarshalJSON implements json.Marshaler.
//func (ex ExAttrs) MarshalJSON() ([]byte, error) {
//	if !s.Valid {
//		return []byte("null"), nil
//	}
//	return json.Marshal(s.String)
//}
//
//// MarshalText implements encoding.TextMarshaler.
//func (ex ExAttrs) MarshalText() ([]byte, error) {
//	if !s.Valid {
//		return []byte{}, nil
//	}
//	return []byte(s.String), nil
//}
//
////// GetBool
////func (ex *ExAttrs) GetBool() bool {
////	return ex.Bool.Bool
////}
////
////// GetFloat64
////func (ex *ExAttrs) GetFloat64() float64 {
////	return ex.Float.Float64
////}
////
////// GetString
////func (ex *ExAttrs) GetString() string {
////	return ex.String.String
////}
////
////// GetInt64
////func (ex *ExAttrs) GetInt64() int64 {
////	return ex.Int.Int64
////}
////
////// GetInt
////func (ex *ExAttrs) GetInt() int {
////	return int(ex.GetInt())
////}
////
////// GetUint
////func (ex *ExAttrs) GetUint() uint {
////	return uint(ex.GetInt())
////}
////
////// GetTime
////func (ex *ExAttrs) GetTime() time.Time {
////	return ex.Time.Time
////}
