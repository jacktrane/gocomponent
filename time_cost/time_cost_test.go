package time_cost

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeCost(t *testing.T) {
	tc := NewTimeCost()
	time.Sleep(1 * time.Second)
	tc.AddPoint("1 point")
	time.Sleep(100 * time.Millisecond)
	tc.AddPoint("2 point")
	time.Sleep(500 * time.Millisecond)
	tc.AddPoint("3 point")
	time.Sleep(2 * time.Second)
	tc.AddPoint("4 point")
	fmt.Println(tc.OutputCostStack(), tc.OutputTotalTime(), tc.OutputCostTime("1 point", "4 point"))
}
