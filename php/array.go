package php

import "github.com/chanwaileung/golang-utils/utils"

func InArray(needle interface{}, haystack []interface{}) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

//对二维的map根据sortName的key值进行排序
func SliceMapSort(values *[]map[string]interface{}, left int, right int, sortName string) {
	first := (*values)[left]

	key := utils.Interface2String(first[sortName])
	p := left
	i, j := left, right
	for i <= j {
		//检索比key小的值，即找到左边部分的数
		for j >= p && utils.Interface2String((*values)[j][sortName]) >= key {
			j--
		}
		if j > p {
			(*values)[p] = (*values)[j]
			p = j
		}

		//检索比key大的值，即找到右部分的数
		for i <= p && utils.Interface2String((*values)[i][sortName]) <= key {
			i++
		}
		if i < p {
			(*values)[p] = (*values)[i]
			p = i
		}
	}

	(*values)[p] = first
	if p-left > 1 {
		SliceMapSort(values, left, p-1, sortName)
	}
	if right-p > 1 {
		SliceMapSort(values, p+1, right, sortName)
	}
}