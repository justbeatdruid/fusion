package util

import (
	"errors"
	"time"
)

type Time struct {
	time.Time `protobuf:"-"`
}

func Now() Time {
	return Time{
		Time: time.Now(),
	}
}

func NewTime(t time.Time) Time {
	return Time{t}
}

const (
	Simple = "2006-01-02 15:04:05"
)

// MarshalJSON implements the json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.Unix() == 0 {
		return []byte("\"\""), nil
	}
	if t.IsZero() {
		return []byte("\"null\""), nil
	}
	if y := t.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(Simple)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, Simple)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Time) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	if string(data) == "\"\"" {
		t.Time = time.Unix(0, 0)
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	var err error
	tm, err := time.Parse(`"`+Simple+`"`, string(data))
	t.Time = tm
	return err
}
