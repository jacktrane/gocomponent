package time_format

import "time"

const (
	OneMin  = 60
	OneHour = OneMin * 60
	OneDay  = OneHour * 24
)

func GetTimestamp() int64 {
	return time.Now().Unix()
}

func GetNanoTimestamp() int64 {
	return time.Now().NanoUnix()
}

func 