package bigqueryGoDate

import (
	"database/sql/driver"
	"fmt"
	"time"

	"cloud.google.com/go/civil"
)

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
			Month: int(v.Month),
			Day:   v.Day,
		}
	default:
		return fmt.Errorf("no se puede convertir %T a Date", value)
	}
	return nil
}
func (d *Date) Scanner(value interface{}) error {
	return d.Scan(value)
}

// Value implementa el interface driver.Valuer para Date
func (d Date) Value() (driver.Value, error) {
	return d.String(), nil
}
func (d Date) Valuer() (driver.Value, error) {
	return d.String(), nil
}
