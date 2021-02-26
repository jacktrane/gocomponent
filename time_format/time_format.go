package time_format

import "time"

const (
	OneMin         = 60
	OneHour        = OneMin * 60
	OneDay         = OneHour * 24
	FullFormatDate = "2006-01-02 15:04:05"
)

const (
	FullFormatDateSimpleYear   = Year
	FullFormatDateSimpleMon    = FullFormatDateSimpleYear + Mon
	FullFormatDateSimpleDay    = FullFormatDateSimpleMon + Day
	FullFormatDateSimpleHour   = FullFormatDateSimpleDay + Hour
	FullFormatDateSimpleMin    = FullFormatDateSimpleHour + Min
	FullFormatDateSimpleSecond = FullFormatDateSimpleMin + Sec
)

const (
	Year = "2006"
	Mon  = "01"
	Day  = "02"
	Hour = "15"
	Min  = "04"
	Sec  = "05"
)

func GetTimestamp() int64 {
	return time.Now().Unix()
}

func GetNanoTimestamp() int64 {
	return time.Now().UnixNano()
}

func GetParseTime(formatStr, date string) (error, *time.Time) {
	lTime, err := time.ParseInLocation(formatStr, date, time.Local)
	return err, &lTime
}

func GetNowTime() *time.Time {
	tTime := time.Now()
	return &tTime
}
