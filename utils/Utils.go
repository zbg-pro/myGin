package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func ContainKey(data map[string]interface{}, key string) bool {
	if _, ok := data[key]; ok {
		// key 存在，value 是对应的值
		return true
	} else {
		// key 不存在
		return false
	}
}

func IsEmptyCollection(listVal interface{}) bool {
	if list, ok := listVal.([]interface{}); ok && list != nil {
		// 如果存在且不为 nil，打印列表内容
		fmt.Println("List is not nil:", list)

		// 判断列表是否为空
		if len(list) == 0 {
			return true
		} else {
			return false
		}
	} else {
		// 如果不存在或为 nil，打印提示信息
		return true
	}
}

func IsEmptyValue(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	// 添加其他类型的判断...
	default:
		// 其他类型暂不处理，视情况添加
		return false
	}
}

func FilterEmptyStrings(inputMap map[string]interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})

	for key, value := range inputMap {
		switch v := value.(type) {
		case string:
			// 排除空字符串
			if v != "" {
				resultMap[key] = value
			}
		default:
			// 对于其他类型，保留
			resultMap[key] = value
		}
	}

	return resultMap
}

// 将对象转换为 map
func StructToMap(obj interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})

	// 获取对象的反射类型和值
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)

	// 遍历对象的字段
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i).Interface()

		// 将字段名及其对应的值放入 map
		resultMap[field.Name] = fieldValue
	}

	return resultMap
}

func ConvertToNumber(str string) interface{} {
	if strings.Contains(str, ".") {
		// 包含小数点，转换为 float64
		floatVal, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil
		}
		return floatVal
	}

	// 不包含小数点，转换为 int64
	intVal, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil
	}
	return intVal
}

// sliceToInClause 将切片转换为 IN 语句字符串
func SliceToInClause(slice interface{}) string {
	var elements []string

	// 使用反射获取切片的元素
	switch v := slice.(type) {
	case []float64:
	case []float32:
	case []int64:
	case []int32:
	case []int16:
	case []int8:
	case []int:
		for _, element := range v {
			elements = append(elements, fmt.Sprintf("%d", element))
		}
	case []string:
		for _, element := range v {
			elements = append(elements, fmt.Sprintf("'%s'", element))
		}
	default:
		// 处理其他切片类型
		val := reflect.ValueOf(slice)
		for i := 0; i < val.Len(); i++ {
			element := val.Index(i).Interface()
			elements = append(elements, fmt.Sprintf("'%v'", element))
		}
	}

	// 使用逗号连接元素，并在两端添加括号
	return "(" + strings.Join(elements, ",") + ")"
}