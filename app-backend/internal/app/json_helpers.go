package app

import (
	"encoding/json"

	"gorm.io/datatypes"
)

func marshalJSON(value any) datatypes.JSON {
	if value == nil {
		return nil
	}
	encoded, _ := json.Marshal(value)
	return datatypes.JSON(encoded)
}
