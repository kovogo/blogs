package table_drvien

import "time"

//GetZeroTimeOfDay 获取目标时间的零点时间
func GetZeroTimeOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(),  0, 0, 0, 0, t.Location())
}

