package schema

import (
	"encoding/json"
	"time"
)

type Reminder struct {
	ID int64
	At time.Time
}

func (r *Reminder) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Reminder) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}
