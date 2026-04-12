package domain

import "time"

var nowFunc = time.Now().UTC

func NowUTC() time.Time {
	return nowFunc()
}
