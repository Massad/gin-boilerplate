package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"regexp"
)

//JSONRaw ...
type JSONRaw json.RawMessage

//Value ...
func (j JSONRaw) Value() (driver.Value, error) {
	byteArr := []byte(j)
	return driver.Value(byteArr), nil
}

//Scan ...
func (j *JSONRaw) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		return error(errors.New("Scan source was not []bytes"))
	}
	err := json.Unmarshal(asBytes, &j)
	if err != nil {
		return error(errors.New("Scan could not unmarshal to []string"))
	}
	return nil
}

//MarshalJSON ...
func (j *JSONRaw) MarshalJSON() ([]byte, error) {
	return *j, nil
}

//UnmarshalJSON ...
func (j *JSONRaw) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

//validateEmail ...
func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}
