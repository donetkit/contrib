package gtime

import "fmt"

const (
	// SecondsPerMinute 定义每分钟的秒数
	SecondsPerMinute = 60
	// SecondsPerHour 定义每小时的秒数
	SecondsPerHour = SecondsPerMinute * 60
	// SecondsPerDay 定义每天的秒数
	SecondsPerDay = SecondsPerHour * 24
)

// ResolveTime 将传入的“秒”解析为3种时间单位
func ResolveTime(seconds int) (day int, hour int, minute int) {
	day = seconds / SecondsPerDay
	hour = seconds / SecondsPerHour
	minute = seconds / SecondsPerMinute
	minute = seconds % SecondsPerMinute
	return
}

// ResolveTimeSecond 将传入的“秒”解析时间单位
func ResolveTimeSecond(seconds int) string {
	day := seconds / SecondsPerDay       //转换天数
	seconds = seconds % SecondsPerDay    //剩余秒数
	hour := seconds / SecondsPerHour     //转换小时
	seconds = seconds % SecondsPerHour   //剩余秒数
	minute := seconds / SecondsPerMinute //转换分钟
	second := seconds % SecondsPerMinute //剩余秒数
	if day > 0 {
		return fmt.Sprintf("%d天%d小时%d分%d秒", day, hour, minute, second)
	} else {
		return fmt.Sprintf("%d小时%d分%d秒", hour, minute, second)
	}
}
