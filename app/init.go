package app

import "time"

func init() {
	// all time now will be in UTC timezone
	time.Local = time.UTC
}
