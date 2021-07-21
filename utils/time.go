package utils

import (
	"fmt"
	"strconv"
	"time"
)

/*
由于php中的date("W")是从周一开始算起的，所以本函数也从周一开始算起
新的一年如果前面有不足一周的，都是算到上一年的最后一周，然后新的一周是从第1周开始算
将返回的格式从只返回周数（int）改成是YYYY-WW（string）的格式
*/
func WeekNumInYear(t time.Time) string {
	yearDay := t.YearDay()		//当前天数（在当年的第几天）
	yearFirstDay := t.AddDate(0, 0, -yearDay+1)
	firstDayInWeek := int(yearFirstDay.Weekday())

	//今年第一周有几天，周日的时候firstWeekDays=0，故只有1天
	firstWeekDays := 1
	if firstDayInWeek != 0 {
		firstWeekDays = 7 - firstDayInWeek + 1
	}

	var week int
	if firstWeekDays < 7 && yearDay < firstWeekDays {
		//当年年初不完整的一周就划到上一年最后一周
		lastYearLastDay, _ := time.ParseInLocation("20060102", strconv.Itoa(t.Year()-1)+"1231", time.Local)
		return WeekNumInYear(lastYearLastDay)
	}else {
		week = (yearDay-firstWeekDays)/7
		if (yearDay-firstWeekDays)%7 != 0 {
			week += 1	//整除则是还在本周内，不整除则需要+1定位到当前周数
		}
	}

	return fmt.Sprintf("%d-%02d", t.Year(), week)
}

/*
获取获取num天的日期,其中includeEndDate为是否包含dateTime当天
num: 获取多少天
dateTime: 结束当天,传空默认是今天的日期
includeEndDate: 是否包含dateTime当天
dataFormat: 返回的日期格式,传空默认是年月日("20060102")的形式
isDesc：是否降序排序.false:升序 true:降序
*/
func GetLastNDays(num int, dateTime string, includeEndDate bool, dateFormat string, isDesc bool) []string {
	var res []string

	if dateFormat == "" {
		dateFormat = "20060102"
	}

	var endTime time.Time
	if dateTime == "" {
		endTime = time.Now()
	} else {
		endTime, _ = time.ParseInLocation(dateFormat, dateTime, time.Local)
	}

	if !includeEndDate {
		endTime = endTime.AddDate(0, 0, -1)
	}

	var calcuTime time.Time
	var diff time.Duration
	if isDesc {
		//降序
		calcuTime = endTime
		diff, _ = time.ParseDuration("-24h")
	}else {
		calcuTime = endTime.AddDate(0, 0, -num+1)
		diff, _ = time.ParseDuration("24h")
	}

	for i := 0; i < num; i++ {
		res = append(res, calcuTime.Format(dateFormat))
		calcuTime = calcuTime.Add(diff)
	}
	return res
}
