package mysql

import (
	"github.com/chanwaileung/golang-utils/utils"
)

//入参必须是struct的指针类型
func GetSqlFieldByStruct(i interface{}) []string {
	return utils.GetFieldByStruct(i, "db")
}

//入参必须是struct的指针类型
func SetTagFieldsStr(i interface{})  string {
	return utils.SetTagFieldsStr(i, "db")
}