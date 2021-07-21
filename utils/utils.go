package utils

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
	"net/url"
)

//通过reflect将带有指定tag的字段拼接成字符串
//入参必须是struct的指针类型
func SetTagFieldsStr(i interface{}, tag string)  string {
	fields := GetFieldByStruct(i, tag)
	return "`" + strings.Join(fields, "`,`") + "`"
}

//入参必须是struct的指针类型
// 示例：fields := GetFieldByStruct(&TestStruct{Id: 0}, "db")
func GetFieldByStruct(i interface{}, tag string) []string {
	t := reflect.TypeOf(i).Elem()
	var fields []string

	for i := 0;i < t.NumField(); i++ {
		if tmp := t.Field(i).Tag.Get(tag); tmp != "" {
			fields = append(fields, t.Field(i).Tag.Get(tag))
		}
	}

	return fields
}

func Interface2String(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case int:
		key = strconv.Itoa(value.(int))
	case uint:
		key = strconv.Itoa(int(value.(uint)))
	case int8:
		key = strconv.Itoa(int(value.(int8)))
	case uint8:
		key = strconv.Itoa(int(value.(uint8)))
	case int16:
		key = strconv.Itoa(int(value.(int16)))
	case uint16:
		key = strconv.Itoa(int(value.(uint16)))
	case int64:
		key = strconv.FormatInt(value.(int64), 10)
	case uint64:
		key = strconv.FormatUint(value.(uint64), 10)
	case float32:
		key = strconv.FormatFloat(float64(value.(float32)), 'f', -1, 64)
	case float64:
		key = strconv.FormatFloat(value.(float64), 'f', -1, 64)
	case rune:
		key = string(value.(rune))
	case bool:
		key = strconv.FormatBool(value.(bool))
	case string:
		key = value.(string)
	default:
		key = fmt.Sprintf("%s", value)
	}
	return key
}

func LessValue(a, b reflect.Value) bool {
	aValue, aNumerical := NumericalValue(a)
	bValue, bNumerical := NumericalValue(b)

	if aNumerical && bNumerical {
		return aValue < bValue
	}

	if !aNumerical && !bNumerical {
		return strings.Compare(a.String(), b.String()) < 0
	}

	return aNumerical && !bNumerical
}

func NumericalValue(value reflect.Value) (float64, bool) {
	switch value.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint()), true

	case reflect.Float32, reflect.Float64:
		return value.Float(), true

	default:
		return 0, false
	}
}

//float类型数据的四舍五入，保留precision位的数字
func FloatRound(val float64, precision uint) (res float64) {
	res, _ = decimal.NewFromFloat(val).Round(int32(precision)).Float64()
	return
}

//如果使用go原生的json处理函数，会将<、>、&等特殊字符转义成Unicode，使用如下方式编码json可以避免这个问题，主要是设置了SetEscapeHTML(false)
func MarshalJsonNotEscapt(data interface{}) ([]byte, error) {
	bf := bytes.NewBuffer([]byte{})
	jsEncoder := jsoniter.ConfigCompatibleWithStandardLibrary.NewEncoder(bf)
	jsEncoder.SetEscapeHTML(false)
	if err := jsEncoder.Encode(data);err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}


func HttpRequest(method string, urlPath string, params map[string]string, header map[string]string, timeout time.Duration, body interface{}) ([]byte, error) {
	//拼接url请求参数
	if strings.ToUpper(method) != "POST" {
		urlParams := url.Values{}
		for k, v := range params {
			urlParams.Add(k, v)
		}
		urlPath = fmt.Sprintf("%s?%s", urlPath, urlParams.Encode())
	}

	//构建body
	bodyJson, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(body)

	request, _ := http.NewRequest(method, urlPath, bytes.NewReader(bodyJson))

	//添加参数
	if strings.ToUpper(method) == "POST" {
		queryParams := request.URL.Query()
		for k, v := range params {
			queryParams.Add(k, v)
		}
		request.URL.RawQuery = queryParams.Encode()
	}

	if header != nil {
		for k, v := range header {
			request.Header.Add(k, v)
		}
	}

	client := &http.Client{Timeout: timeout * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		response.Body.Close()
	}()

	result, _ := ioutil.ReadAll(response.Body)
	return result, nil
}
