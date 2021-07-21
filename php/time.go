package php

import (
	"fmt"
	"strconv"
	"time"
)

//与php的microtime一致，主要为了获取时间戳小数点后
//如果getAsFloat为true，则返回一个精确到万分位（4位小数）的时间戳，否则
func MicroTime(getAsFloat bool) string {
	localTime := time.Now()
	nano := strconv.FormatInt(localTime.UnixNano(), 10)[10:18]
	if getAsFloat {
		Fnano, err := strconv.ParseFloat("0."+nano, 64)
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%.4f", float64(localTime.Unix()) + Fnano)
	}else {
		return fmt.Sprintf("0.%v %v", nano, localTime.Unix())
	}
}
