package time

import (
	"encoding/json"
	"time"
)

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.Unix())
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var i int64

	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	t.Time = time.Unix(i, 0)
	return nil
}
