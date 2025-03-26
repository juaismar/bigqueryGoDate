package bigqueryGoDate

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"

	"cloud.google.com/go/civil"
)

// A Date represents a date (year, month, day).
//
// This type does not include location information, and therefore does not
// describe a unique 24-hour timespan.
type Date struct {
	Year  int        // Year (e.g., 2014).
	Month time.Month // Month of the year (January = 1, ...).
	Day   int        // Day of the month, starting at 1.
}

// DateOf returns the Date in which a time occurs in that time's location.
func DateOf(t time.Time) Date {
	var d Date
	d.Year, d.Month, d.Day = t.Date()
	return d
}

// ParseDate parses a string in RFC3339 full-date format and returns the date value it represents.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return Date{}, err
	}
	return DateOf(t), nil
}

// String returns the date in RFC3339 full-date format.
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// IsValid reports whether the date is valid.
func (d Date) IsValid() bool {
	return DateOf(d.In(time.UTC)) == d
}

// In returns the time corresponding to time 00:00:00 of the date in the location.
//
// In is always consistent with time.Date, even when time.Date returns a time
// on a different day. For example, if loc is America/Indiana/Vincennes, then both
//
//	time.Date(1955, time.May, 1, 0, 0, 0, 0, loc)
//
// and
//
//	civil.Date{Year: 1955, Month: time.May, Day: 1}.In(loc)
//
// return 23:00:00 on April 30, 1955.
//
// In panics if loc is nil.
func (d Date) In(loc *time.Location) time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, loc)
}

// AddDays returns the date that is n days in the future.
// n can also be negative to go into the past.
func (d Date) AddDays(n int) Date {
	return DateOf(d.In(time.UTC).AddDate(0, 0, n))
}

// DaysSince returns the signed number of days between the date and s, not including the end day.
// This is the inverse operation to AddDays.
func (d Date) DaysSince(s Date) (days int) {
	// We convert to Unix time so we do not have to worry about leap seconds:
	// Unix time increases by exactly 86400 seconds per day.
	deltaUnix := d.In(time.UTC).Unix() - s.In(time.UTC).Unix()
	return int(deltaUnix / 86400)
}

// Before reports whether d occurs before d2.
func (d Date) Before(d2 Date) bool {
	if d.Year != d2.Year {
		return d.Year < d2.Year
	}
	if d.Month != d2.Month {
		return d.Month < d2.Month
	}
	return d.Day < d2.Day
}

// After reports whether d occurs after d2.
func (d Date) After(d2 Date) bool {
	return d2.Before(d)
}

// Compare compares d and d2. If d is before d2, it returns -1;
// if d is after d2, it returns +1; otherwise it returns 0.
func (d Date) Compare(d2 Date) int {
	if d.Before(d2) {
		return -1
	} else if d.After(d2) {
		return +1
	}
	return 0
}

// IsZero reports whether date fields are set to their default value.
func (d Date) IsZero() bool {
	return (d.Year == 0) && (int(d.Month) == 0) && (d.Day == 0)
}

// MarshalText implements the encoding.TextMarshaler interface.
// The output is the result of d.String().
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The date is expected to be a string in a format accepted by ParseDate.
func (d *Date) UnmarshalText(data []byte) error {
	var err error
	*d, err = ParseDate(string(data))
	return err
}

// Scan implementa el interface sql.Scanner para Date
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		*d = Date{}
		return nil
	}

	switch v := value.(type) {
	case string:
		parsed, err := ParseDate(v)
		if err != nil {
			return err
		}
		*d = parsed
	case time.Time:
		*d = DateOf(v)
	case civil.Date:
		*d = Date{
			Year:  v.Year,
			Month: v.Month,
			Day:   v.Day,
		}
	default:
		return fmt.Errorf("no se puede convertir %T a Date", value)
	}
	return nil
}

// Value implementa el interface driver.Valuer para Date
func (d Date) Value() (driver.Value, error) {
	return d.String(), nil
}

// A Time represents a time with nanosecond precision.
//
// This type does not include location information, and therefore does not
// describe a unique moment in time.
//
// This type exists to represent the TIME type in storage-based APIs like BigQuery.
// Most operations on Times are unlikely to be meaningful. Prefer the DateTime type.
type Time struct {
	Hour       int // The hour of the day in 24-hour format; range [0-23]
	Minute     int // The minute of the hour; range [0-59]
	Second     int // The second of the minute; range [0-59]
	Nanosecond int // The nanosecond of the second; range [0-999999999]
}

// TimeOf returns the Time representing the time of day in which a time occurs
// in that time's location. It ignores the date.
func TimeOf(t time.Time) Time {
	var tm Time
	tm.Hour, tm.Minute, tm.Second = t.Clock()
	tm.Nanosecond = t.Nanosecond()
	return tm
}

// ParseTime parses a string and returns the time value it represents.
// ParseTime accepts an extended form of the RFC3339 partial-time format. After
// the HH:MM:SS part of the string, an optional fractional part may appear,
// consisting of a decimal point followed by one to nine decimal digits.
// (RFC3339 admits only one digit after the decimal point).
func ParseTime(s string) (Time, error) {
	t, err := time.Parse("15:04:05.999999999", s)
	if err != nil {
		return Time{}, err
	}
	return TimeOf(t), nil
}

// String returns the date in the format described in ParseTime. If Nanoseconds
// is zero, no fractional part will be generated. Otherwise, the result will
// end with a fractional part consisting of a decimal point and nine digits.
func (t Time) String() string {
	s := fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
	if t.Nanosecond == 0 {
		return s
	}
	return s + fmt.Sprintf(".%09d", t.Nanosecond)
}

// IsValid reports whether the time is valid.
func (t Time) IsValid() bool {
	// Construct a non-zero time.
	tm := time.Date(2, 2, 2, t.Hour, t.Minute, t.Second, t.Nanosecond, time.UTC)
	return TimeOf(tm) == t
}

// IsZero reports whether time fields are set to their default value.
func (t Time) IsZero() bool {
	return (t.Hour == 0) && (t.Minute == 0) && (t.Second == 0) && (t.Nanosecond == 0)
}

// Before reports whether t occurs before t2.
func (t Time) Before(t2 Time) bool {
	if t.Hour != t2.Hour {
		return t.Hour < t2.Hour
	}
	if t.Minute != t2.Minute {
		return t.Minute < t2.Minute
	}
	if t.Second != t2.Second {
		return t.Second < t2.Second
	}

	return t.Nanosecond < t2.Nanosecond
}

// After reports whether t occurs after t2.
func (t Time) After(t2 Time) bool {
	return t2.Before(t)
}

// Compare compares t and t2. If t is before t2, it returns -1;
// if t is after t2, it returns +1; otherwise it returns 0.
func (t Time) Compare(t2 Time) int {
	if t.Before(t2) {
		return -1
	} else if t.After(t2) {
		return +1
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
// The output is the result of t.String().
func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The time is expected to be a string in a format accepted by ParseTime.
func (t *Time) UnmarshalText(data []byte) error {
	var err error
	*t, err = ParseTime(string(data))
	return err
}

// Value implements the database/sql/driver Valuer interface.
func (t Time) Value() (driver.Value, error) {
	return t.String(), nil
}

// Scan implements the database/sql Scanner interface.
func (t *Time) Scan(v any) error {
	switch vt := v.(type) {
	case time.Time:
		*t = TimeOf(vt)
	case *time.Time:
		if vt != nil {
			*t = TimeOf(*vt)
		}
	case string:
		var err error
		*t, err = ParseTime(vt)
		return err
	case *string:
		var err error
		if vt != nil {
			*t, err = ParseTime(*vt)
		}
		return err
	case []byte:
		var err error
		*t, err = ParseTime(string(vt))
		return err
	case *[]byte:
		var err error
		if vt != nil {
			*t, err = ParseTime(string(*vt))
		}
		return err
	case civil.Time:
		*t = Time{
			Hour:       vt.Hour,
			Minute:     vt.Minute,
			Second:     vt.Second,
			Nanosecond: vt.Nanosecond,
		}
	default:
		return fmt.Errorf("unsupported scan type for Time: %T", v)
	}
	return nil
}

// A DateTime represents a date and time.
//
// This type does not include location information, and therefore does not
// describe a unique moment in time.
type DateTime struct {
	Date Date
	Time Time
}

// Note: We deliberately do not embed Date into DateTime, to avoid promoting AddDays and Sub.

// DateTimeOf returns the DateTime in which a time occurs in that time's location.
func DateTimeOf(t time.Time) DateTime {
	return DateTime{
		Date: DateOf(t),
		Time: TimeOf(t),
	}
}

// ParseDateTime parses a string and returns the DateTime it represents.
// ParseDateTime accepts a variant of the RFC3339 date-time format that omits
// the time offset but includes an optional fractional time, as described in
// ParseTime. Informally, the accepted format is
//
//	YYYY-MM-DDTHH:MM:SS[.FFFFFFFFF]
//
// where the 'T' may be a lower-case 't'.
func ParseDateTime(s string) (DateTime, error) {
	t, err := time.Parse("2006-01-02T15:04:05.999999999", s)
	if err != nil {
		t, err = time.Parse("2006-01-02t15:04:05.999999999", s)
		if err != nil {
			return DateTime{}, err
		}
	}
	return DateTimeOf(t), nil
}

// String returns the date in the format described in ParseDate.
func (dt DateTime) String() string {
	return dt.Date.String() + "T" + dt.Time.String()
}

// IsValid reports whether the datetime is valid.
func (dt DateTime) IsValid() bool {
	return dt.Date.IsValid() && dt.Time.IsValid()
}

// In returns the time corresponding to the DateTime in the given location.
//
// If the time is missing or ambigous at the location, In returns the same
// result as time.Date. For example, if loc is America/Indiana/Vincennes, then
// both
//
//	time.Date(1955, time.May, 1, 0, 30, 0, 0, loc)
//
// and
//
//	civil.DateTime{
//	    civil.Date{Year: 1955, Month: time.May, Day: 1}},
//	    civil.Time{Minute: 30}}.In(loc)
//
// return 23:30:00 on April 30, 1955.
//
// In panics if loc is nil.
func (dt DateTime) In(loc *time.Location) time.Time {
	return time.Date(dt.Date.Year, dt.Date.Month, dt.Date.Day, dt.Time.Hour, dt.Time.Minute, dt.Time.Second, dt.Time.Nanosecond, loc)
}

// Before reports whether dt occurs before dt2.
func (dt DateTime) Before(dt2 DateTime) bool {
	return dt.In(time.UTC).Before(dt2.In(time.UTC))
}

// After reports whether dt occurs after dt2.
func (dt DateTime) After(dt2 DateTime) bool {
	return dt2.Before(dt)
}

// Compare compares dt and dt2. If dt is before dt2, it returns -1;
// if dt is after dt2, it returns +1; otherwise it returns 0.
func (dt DateTime) Compare(dt2 DateTime) int {
	return dt.In(time.UTC).Compare(dt2.In(time.UTC))
}

// IsZero reports whether datetime fields are set to their default value.
func (dt DateTime) IsZero() bool {
	return dt.Date.IsZero() && dt.Time.IsZero()
}

// MarshalText implements the encoding.TextMarshaler interface.
// The output is the result of dt.String().
func (dt DateTime) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The datetime is expected to be a string in a format accepted by ParseDateTime
func (dt *DateTime) UnmarshalText(data []byte) error {
	var err error
	*dt, err = ParseDateTime(string(data))
	return err
}

// Value implements the database/sql/driver Valuer interface.
func (dt DateTime) Value() (driver.Value, error) {
	return dt.String(), nil
}

// Scan implements the database/sql Scanner interface.
func (dt *DateTime) Scan(v any) error {
	switch vt := v.(type) {
	case time.Time:
		*dt = DateTimeOf(vt)
	case *time.Time:
		if vt != nil {
			*dt = DateTimeOf(*vt)
		}
	case civil.DateTime:
		*dt = DateTime{
			Date: Date{
				Year:  vt.Date.Year,
				Month: vt.Date.Month,
				Day:   vt.Date.Day,
			},
			Time: Time{
				Hour:       vt.Time.Hour,
				Minute:     vt.Time.Minute,
				Second:     vt.Time.Second,
				Nanosecond: vt.Time.Nanosecond,
			},
		}
	case string:
		var err error
		*dt, err = ParseDateTime(vt)
		return err
	case *string:
		var err error
		if vt != nil {
			*dt, err = ParseDateTime(*vt)
		}
		return err
	case []byte:
		var err error
		*dt, err = ParseDateTime(string(vt))
		return err
	case *[]byte:
		var err error
		if vt != nil {
			*dt, err = ParseDateTime(string(*vt))
		}
		return err
	default:
		fmt.Printf("Type: %v\n", reflect.TypeOf(v))
		fmt.Printf("Value: %v\n", v)
		return fmt.Errorf("unsupported scan type for DateTime: %T", v)
	}
	return nil
}
