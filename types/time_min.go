package types

import (
	"context"
	"github.com/sanches1984/gopkg-errors"
	"strconv"
	"time"
)

var TimeMinParseError = errors.Internal.Err(context.Background(), `TimeParseError: should be a integer`)
var TimeMinMaxError = errors.Internal.Err(context.Background(), `TimeParseError: should be a integer less then 1440`)

func NewTimeMin(t *time.Time) TimeMin {
	return TimeMin{t: t}
}

func NewNotEmptyTimeMinFromString(s string) (TimeMin, error) {
	if s == "" {
		return TimeMin{}, errors.Internal.Err(context.Background(), "TimeParseError: should be not empty")
	}
	t, err := NewTimeMinFromString(s)
	if err != nil {
		return TimeMin{}, err
	}
	return t, nil
}

func NewTimeMinFromString(s string) (TimeMin, error) {
	t := TimeMin{}
	if s == "" {
		return t, nil
	}
	if err := t.decodeStr(s); err != nil {
		return t, err
	}
	return t, nil
}

func NewTimeMinFromInt(min int16) (TimeMin, error) {
	t := TimeMin{}
	if min == 0 {
		return t, nil
	}
	if err := t.decodeInt(min); err != nil {
		return t, err
	}
	return t, nil
}

type TimeMin struct {
	t *time.Time
}

func (t TimeMin) MarshalJSON() ([]byte, error) {
	if t.t == nil {
		return []byte(`0`), nil
	}
	return []byte(strconv.Itoa(t.t.Hour()*60 + t.t.Minute())), nil
}

func (t *TimeMin) UnmarshalJSON(b []byte) error {
	s := string(b)
	i, err := strconv.Atoi(s)
	if err != nil {
		return TimeMinParseError.WithPayloadKV("actual", s)
	}
	return t.decodeInt(int16(i))
}

func (t *TimeMin) decodeStr(s string) error {
	if len(s) != 5 {
		return TimeParseError.WithPayloadKV("actual", s)
	}
	ret, err := time.Parse(TimeLayout, s)
	if err != nil {
		return err
	}
	t.t = &ret
	return nil
}

func (t *TimeMin) decodeInt(min int16) error {
	if min > 1440 {
		return TimeMinMaxError.WithPayloadKV("actual", min)
	}
	dt := time.Date(0, 0, 0, int(min)/60, int(min)%60, 0, 0, time.UTC)
	t.t = &dt
	return nil
}

func (t TimeMin) ToHHMM() string {
	if t.t == nil {
		return ""
	}
	return t.t.Format(TimeLayout)
}

func (t TimeMin) ToMin() int16 {
	if t.t == nil {
		return 0
	}
	return int16(t.t.Hour()*60 + t.t.Minute())
}

func (t TimeMin) ToTime() time.Time {
	if t.t == nil {
		return time.Time{}
	}
	return *t.t
}

func (t TimeMin) ToDuration() time.Duration {
	if t.t == nil {
		return 0
	}
	return time.Minute * time.Duration(t.ToMin())
}

func (t TimeMin) After(dt time.Time) bool {
	if t.t == nil {
		return false
	}
	return int16(dt.Hour()*60+dt.Minute()) < t.ToMin()
}

func (t TimeMin) Before(dt time.Time) bool {
	if t.t == nil {
		return true
	}
	return int16(dt.Hour()*60+dt.Minute()) >= t.ToMin()
}

func (t TimeMin) Equal(dt *time.Time) bool {
	if t.t == nil {
		return dt == nil
	}
	if dt == nil {
		return t.t == nil
	}
	return t.t.Equal(*dt)
}
