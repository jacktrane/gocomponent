package time_cost

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	NS  = "ns"
	MMS = "mms"
	MS  = "ms"
	S   = "s"
	M   = "m"
)

type TimeCost struct {
	unit          string
	paths         []string
	pathTimePoint sync.Map
	t             time.Time
}

func NewTimeCost() *TimeCost {
	return &TimeCost{
		unit:  MS,
		paths: make([]string, 0, 10),
		t:     time.Now(),
	}
}

func (tc *TimeCost) SetUnit(unit string) error {
	if len(tc.paths) != 0 {
		return errors.New("timecost has point")
	}
	tc.unit = unit
	return nil
}

func formatPointKey(pointName string, index int) string {
	return pointName + strconv.Itoa(index)
}

// 打点
func (tc *TimeCost) AddPoint(pointName string) {
	tc.paths = append(tc.paths, pointName)
	tc.pathTimePoint.Store(formatPointKey(pointName, len(tc.paths)-1), time.Since(tc.t))
	tc.t = time.Now()
}

func (tc *TimeCost) outputCost(d time.Duration) int64 {
	switch tc.unit {
	case NS:
		return d.Nanoseconds()
	case MMS:
		return d.Microseconds()
	case MS:
		return d.Milliseconds()
	case S:
		return int64(d.Seconds())
	case M:
		return int64(d.Minutes())
	default:
		return int64(d.Milliseconds())
	}
}

// 输出链路
func (tc *TimeCost) OutputCostStack() string {
	var totalCost time.Duration
	arrPointCost := make([]string, len(tc.paths)+2)
	arrPointCost[0] = "Begin"
	for index, pointName := range tc.paths {
		fmt.Println(pointName)
		duration, ok := tc.pathTimePoint.Load(formatPointKey(pointName, index))
		if ok {
			timeDuration := duration.(time.Duration)
			totalCost += timeDuration
			arrPointCost[index+1] = fmt.Sprintf("(%s %d%s)", pointName, tc.outputCost(timeDuration), tc.unit)
		}
	}

	arrPointCost[len(tc.paths)+2-1] = "End"
	return "TotalCost:" + strconv.Itoa(int(tc.outputCost(totalCost))) + ";Detail:" + strings.Join(arrPointCost, "=>")
}

// 输出两段路径之间的耗时
func (tc *TimeCost) OutputCostTime(bgnName, endName string) int64 {
	var bgnTime, endTime time.Duration = 0, 0
	for index, pointName := range tc.paths {
		if bgnTime != 0 && endTime != 0 {
			break
		}
		if pointName == bgnName {
			if duration, ok := tc.pathTimePoint.Load(formatPointKey(pointName, index)); ok {
				bgnTime = duration.(time.Duration)
			}
		}
		if pointName == endName {
			if duration, ok := tc.pathTimePoint.Load(formatPointKey(pointName, index)); ok {
				endTime = duration.(time.Duration)
			}
		}

	}
	return tc.outputCost(endTime - bgnTime)
}

// 输出总耗时
func (tc *TimeCost) OutputTotalTime() int64 {
	var totalCost time.Duration
	for index, pointName := range tc.paths {
		duration, ok := tc.pathTimePoint.Load(formatPointKey(pointName, index))
		if ok {
			timeDuration := duration.(time.Duration)
			totalCost += timeDuration
		}
	}
	return tc.outputCost(totalCost)
}
