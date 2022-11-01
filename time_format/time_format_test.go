package time_format

import (
	"testing"
	"time"
)

func TestSpecDayBgnAndEndTime(t *testing.T) {
	t1, _ := time.ParseInLocation(FullFormatDate, "2022-11-01 14:40:00", time.Local)
	
	tB, tE := SpecDayBgnAndEndTime(&t1)
	if tB.Format(FullFormatDate) != "2022-11-01 00:00:00" {
		t.Fatal("time bgn error")
	}
	if tE.Format(FullFormatDate) != "2022-11-01 23:59:59" {
		t.Fatal("time end error")
	}
}