package time_cost

import "sync"

var (
	MMS = "mms"
	MS  = "ms"
	S   = "s"
	M   = "m"
)

type TimeCost struct {
	unit          string
	paths         []string
	pathTimePoint sync.Map
}

func NewTimeCost() *TimeCost {
	return &TimeCost{
		unit: MS,
	}
}

func (tc *TimeCost) SetUnit(unit string) error {
	tc.unit = unit
	return nil
}

// TODO 打点
func (tc *TimeCost) AddPoint(pointName string) {

}

// TODO 输出链路
func (tc *TimeCost) OutputCostStack() string {
	return ""
}

// TODO 输出两段路径之间的耗时
func (tc *TimeCost) OutputCostTime(bgnName, endName string) int64 {
	return 0
}

// TODO 输出总耗时
func (tc *TimeCost) OutputTotalTime() int64 {
	return 0
}
