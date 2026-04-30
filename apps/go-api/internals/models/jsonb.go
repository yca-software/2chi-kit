package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

func JSONBValue(v any) (driver.Value, error) {
	return json.Marshal(v)
}

func JSONBScan(value any, dest any) error {
	j, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(j, dest)
}
