package dao

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	log2 "myGin/log"
	"myGin/utils"
	"reflect"
	"regexp"
	"strings"
)

type QueryModel interface {
	TableName() string
}

func init() {
	log2.InitLogger()
}

// 通用查询
// 通用查询
func Query(db *gorm.DB, modelType QueryModel, obj map[string]interface{}, result interface{}) (tx *gorm.DB) {
	if db == nil {
		var err error
		db, err = gorm.Open(sqlite.Open("../sqllite/sqlLite-database.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
		if err != nil {
			log.Fatal("error:", err)
		}
	}
	conditions := getConditions(obj, modelType)
	return db.Table(modelType.TableName()).Where(conditions[0], conditions[1:]...).Find(result)
}

func getConditions(paramMap map[string]interface{}, modelType QueryModel) []interface{} {
	structType := reflect.TypeOf(modelType)
	paramMap = utils.FilterEmptyStrings(paramMap)

	var conditions []interface{}
	var sql = " 1=1 "
	var params []interface{}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		isString := field.Type.Kind() == reflect.String

		columnName := getColumnName(field)
		jsonProName := getJsonProName(field)
		fmt.Println(columnName)

		if utils.IsEmptyValue(columnName) {
			continue
		}

		//=
		if utils.ContainKey(paramMap, jsonProName) {
			fieldValue := paramMap[jsonProName]
			sql = sql + " and " + columnName + "=? "
			params = append(params, paramMap[jsonProName])

			if isString {
				params = append(params, fieldValue.(string))
			} else {
				fieldValue = utils.ConvertToNumber(fieldValue.(string))
				params = append(params, fieldValue)
			}

		}

		// like
		if utils.ContainKey(paramMap, jsonProName+"Like") {
			fieldValue := "%" + paramMap[jsonProName+"Like"].(string) + "%"
			sql = sql + " and " + columnName + " like ? "

			params = append(params, fieldValue)

		}

		//!=
		if utils.ContainKey(paramMap, jsonProName+"Nq") {
			fieldValue := paramMap[jsonProName+"Nq"]
			sql = sql + " and " + columnName + "!=? "

			if isString {
				params = append(params, fieldValue.(string))
			} else {
				fieldValue = utils.ConvertToNumber(fieldValue.(string))
				params = append(params, fieldValue)
			}

		}

		// > <
		if utils.ContainKey(paramMap, jsonProName+"Start") {
			fieldValue := paramMap[jsonProName+"Start"]
			sql = sql + " and " + columnName + ">? "

			if isString {
				params = append(params, fieldValue.(string))
			} else {
				fieldValue = utils.ConvertToNumber(fieldValue.(string))
				params = append(params, fieldValue)
			}

		}
		if utils.ContainKey(paramMap, jsonProName+"End") {
			fieldValue := paramMap[jsonProName+"End"]
			sql = sql + " and " + columnName + "<? "

			if isString {
				params = append(params, fieldValue.(string))
			} else {
				fieldValue = utils.ConvertToNumber(fieldValue.(string))
				params = append(params, fieldValue)
			}

		}

		//>= <=
		if utils.ContainKey(paramMap, jsonProName+"Min") {
			fieldValue := paramMap[jsonProName+"Min"]
			sql = sql + " and " + columnName + ">=? "

			if isString {
				params = append(params, fieldValue.(string))
			} else {
				fieldValue = utils.ConvertToNumber(fieldValue.(string))
				params = append(params, fieldValue)
			}

		}

		//in()
		if utils.ContainKey(paramMap, jsonProName+"List") {
			fieldValue := paramMap[jsonProName+"List"]
			if utils.IsEmptyCollection(fieldValue) {
				continue
			}

			sql = sql + " and " + columnName + " in " + utils.SliceToInClause(fieldValue.([]interface{}))
		}
		//not in()
		if utils.ContainKey(paramMap, jsonProName+"NqList") {
			fieldValue := paramMap[jsonProName+"NqList"]
			if utils.IsEmptyCollection(fieldValue) {
				continue
			}

			sql = sql + " and " + columnName + " not in " + utils.SliceToInClause(fieldValue.([]interface{}))
		}
	}

	fmt.Println(sql)
	fmt.Println(params)
	conditions = append(conditions, sql)

	for i := 0; i < len(params); i++ {
		conditions = append(conditions, params[i])
	}

	return conditions
}

func getJsonProName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	return strings.TrimSpace(tag)
}

func getColumnName(field reflect.StructField) string {
	// 定义正则表达式
	reg := regexp.MustCompile(`column:(\w+)`)
	// 获取字段名称
	gormTag := field.Tag.Get("gorm")
	if gormTag == "" {
		return ""
	}
	// 使用正则表达式提取列名称
	matches := reg.FindStringSubmatch(gormTag)
	if len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}
