package dao

import (
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"modernc.org/mathutil"
	log2 "myGin/log"
	"myGin/utils"
	"reflect"
	"regexp"
	"strings"
)

type QueryModel interface {
}

func init() {
	log2.InitLogger()
}

func Select(db *gorm.DB, modelType QueryModel, dest interface{}) error {
	return SelectPage(db, modelType, dest, nil)
}

func SelectPage(db *gorm.DB, modelType QueryModel, dest interface{}, orderByFieldNameMap map[string]bool) error {
	paramMap := utils.StructToMap(modelType)
	tableName := utils.CamelToSnake(GetModelType(dest).Name())
	return SelectPageByParamMap(db, paramMap, modelType, tableName, dest, orderByFieldNameMap)
}

func SelectByParamMap(db *gorm.DB, paramMap map[string]interface{}, modelCustomType QueryModel, tableName string, dest interface{}) error {
	return SelectPageByParamMap(db, paramMap, modelCustomType, tableName, dest, nil)
}

var ErrUnsupportedDataType = errors.New("unsupported data type")

func SelectPageByParamMap(db *gorm.DB, paramMap map[string]interface{}, modelCustomType QueryModel, tableName string, dest interface{}, orderByFieldNameMap map[string]bool) error {
	if db == nil {
		var err error
		db, err = gorm.Open(sqlite.Open("../sqllite/sqlLite-database.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
		if err != nil {
			log.Fatal("error:", err)
		}
	}
	if dest == nil {
		return fmt.Errorf("SelectPageByParamMap %w: %+v", ErrUnsupportedDataType, dest)
	}

	if strings.TrimSpace(tableName) == "" {
		tableName = utils.CamelToSnake(GetModelType(dest).Name())
	}
	conditions := GetConditions(paramMap, modelCustomType, orderByFieldNameMap)
	db.Table(tableName).Where(conditions[0], conditions[1:]...).Find(dest)
	return nil
}

// 通用查询
func SelectPageList(db *gorm.DB, modelType QueryModel, paramMap map[string]interface{}, result interface{}, orderByFieldNames map[string]bool) error {
	if db == nil {
		var err error
		db, err = gorm.Open(sqlite.Open("../sqllite/sqlLite-database.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
		if err != nil {
			log.Fatal("error:", err)
		}
	}
	conditions := GetConditions(paramMap, modelType, orderByFieldNames)
	db.Table(utils.CamelToSnake(GetModelType(modelType).Name())).Where(conditions[0], conditions[1:]...).Find(result)
	return nil
}

func SelectById(db *gorm.DB, modelType QueryModel, id interface{}, result interface{}) (tx *gorm.DB) {
	if db == nil {
		var err error
		db, err = gorm.Open(sqlite.Open("../sqllite/sqlLite-database.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
		if err != nil {
			log.Fatal("error:", err)
		}
	}

	paramMap := map[string]interface{}{}
	idName := GetPrimaryKeyJsonName(modelType)
	paramMap[idName] = id

	//给一个json获取到到map
	conditions := GetConditions(paramMap, modelType, nil)

	db.Table(utils.CamelToSnake(GetModelType(modelType).Name())).Where(conditions[0], conditions[1:]...).Find(result)
	return nil
}

func GetConditions(paramMap map[string]interface{}, modelType QueryModel, orderByFieldNames map[string]bool) []interface{} {
	structType := reflect.TypeOf(modelType)
	paramMap = utils.FilterMapEmptyValues(paramMap)

	pageIndex := GetPageIndex(paramMap)
	pageSize := GetPageSize(paramMap)
	orderBySql := GetOrderBySql(modelType, orderByFieldNames, pageIndex, pageSize)
	fmt.Println("orderBySql: ", orderBySql)
	var conditions []interface{}

	var sql = " 1=1 "
	var params []interface{}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		isString := field.Type.Kind() == reflect.String

		fieldName := field.Name
		columnName := GetColumnName(field)
		jsonProName := GetJsonProName(field)

		jsonSql, jsonParam := getWhereParamAndSql(jsonProName, paramMap, columnName, isString)
		columnSql, columnParam := getWhereParamAndSql(columnName, paramMap, columnName, isString)
		filedSql, filedParam := getWhereParamAndSql(fieldName, paramMap, columnName, isString)

		sql = sql + jsonSql + columnSql + filedSql
		if len(jsonParam) > 0 {
			params = append(params, jsonParam...)
		}
		if len(columnParam) > 0 {
			params = append(params, columnParam...)
		}
		if len(filedParam) > 0 {
			params = append(params, filedParam...)
		}

	}

	sql = sql + orderBySql

	conditions = append(conditions, sql)

	for i := 0; i < len(params); i++ {
		conditions = append(conditions, params[i])
	}

	log.Println("whereSql: ", sql)
	log.Println("whereParam: ", utils.ToJSON(params))
	log.Println("conditions:" + utils.ToJSON(conditions))
	return conditions
}

func getWhereParamAndSql(paramKeyName string, paramMap map[string]interface{}, columnName string, isString bool) (string, []interface{}) {
	var sql = ""
	var params []interface{}

	if paramKeyName == "" {
		return sql, params
	}

	//=
	if utils.ContainKey(paramMap, paramKeyName) {
		fieldValue := paramMap[paramKeyName]
		sql = sql + " and " + columnName + "=? "
		delete(paramMap, paramKeyName)
		if isString {
			params = append(params, fieldValue.(string))
		} else {
			fieldValue = utils.ConvertToNumber(fmt.Sprintf("%v", fieldValue))
			params = append(params, fieldValue)
		}

	}

	// like
	if utils.ContainKey(paramMap, paramKeyName+"Like") {
		fieldValue := "%" + paramMap[paramKeyName+"Like"].(string) + "%"
		sql = sql + " and " + columnName + " like ? "
		delete(paramMap, paramKeyName+"Like")
		params = append(params, fieldValue)
	}

	//!=
	if utils.ContainKey(paramMap, paramKeyName+"Nq") {
		fieldValue := paramMap[paramKeyName+"Nq"]
		sql = sql + " and " + columnName + "!=? "
		delete(paramMap, paramKeyName+"Nq")
		if isString {
			params = append(params, fieldValue.(string))
		} else {
			fieldValue = utils.ConvertToNumber(fmt.Sprintf("%v", fieldValue))
			params = append(params, fieldValue)
		}

	}

	// > <
	if utils.ContainKey(paramMap, paramKeyName+"Start") {
		fieldValue := paramMap[paramKeyName+"Start"]
		sql = sql + " and " + columnName + ">? "
		delete(paramMap, paramKeyName+"Start")
		if isString {
			params = append(params, fieldValue.(string))
		} else {
			fieldValue = utils.ConvertToNumber(fmt.Sprintf("%v", fieldValue))
			params = append(params, fieldValue)
		}

	}
	if utils.ContainKey(paramMap, paramKeyName+"End") {
		fieldValue := paramMap[paramKeyName+"End"]
		sql = sql + " and " + columnName + "<? "
		delete(paramMap, paramKeyName+"End")
		if isString {
			params = append(params, fieldValue.(string))
		} else {
			fieldValue = utils.ConvertToNumber(fmt.Sprintf("%v", fieldValue))
			params = append(params, fieldValue)
		}

	}

	//>= <=
	if utils.ContainKey(paramMap, paramKeyName+"Min") {
		fieldValue := paramMap[paramKeyName+"Min"]
		sql = sql + " and " + columnName + ">=? "
		delete(paramMap, paramKeyName+"Min")
		if isString {
			params = append(params, fieldValue.(string))
		} else {
			fieldValue = utils.ConvertToNumber(fmt.Sprintf("%v", fieldValue))
			params = append(params, fieldValue)
		}
	}
	if utils.ContainKey(paramMap, paramKeyName+"Max") {
		fieldValue := paramMap[paramKeyName+"Max"]
		sql = sql + " and " + columnName + "<=? "
		delete(paramMap, paramKeyName+"Max")
		if isString {
			params = append(params, fieldValue.(string))
		} else {
			fieldValue = utils.ConvertToNumber(fmt.Sprintf("%v", fieldValue))
			params = append(params, fieldValue)
		}

	}

	//in()
	if utils.ContainKey(paramMap, paramKeyName+"List") {
		fieldValue := paramMap[paramKeyName+"List"]
		if !utils.IsEmptyCollection(fieldValue) {
			delete(paramMap, paramKeyName+"List")
			// 生成占位符字符串 (?, ?, ?)
			arr := fieldValue.([]interface{})
			if len(arr) > 0 {
				placeholders := make([]string, len(arr))
				for i := range placeholders {
					placeholders[i] = "?"
				}
				for i := 0; i < len(arr); i++ {
					params = append(params, arr[i])
				}

				sql = sql + " and " + columnName + " in (" + strings.Join(placeholders, ",") + ") "
			}
			//sql = sql + " and " + columnName + " in " + utils.SliceToInClause(fieldValue.([]interface{}))
		}

	}
	//not in()
	if utils.ContainKey(paramMap, paramKeyName+"NqList") {
		fieldValue := paramMap[paramKeyName+"NqList"]
		if !utils.IsEmptyCollection(fieldValue) {
			delete(paramMap, paramKeyName+"NqList")

			// 生成占位符字符串 (?, ?, ?)
			arr := fieldValue.([]interface{})
			if len(arr) > 0 {
				placeholders := make([]string, len(arr))
				for i := range placeholders {
					placeholders[i] = "?"
				}
				for i := 0; i < len(arr); i++ {
					params = append(params, arr[i])
				}

				sql = sql + " and " + columnName + " not in (" + strings.Join(placeholders, ",") + ") "
			}
			//sql = sql + " and " + columnName + " not in " + utils.SliceToInClause(fieldValue.([]interface{}))
		}
	}

	return sql, params
}

func GetOrderBySql(modelType QueryModel, orderFiledMap map[string]bool, pageIndex int64, pageSize int64) string {
	//param.put("offset", (pageIndex - 1) * pageSize);
	//            param.put("pageSize", pageSize);

	limitSql := fmt.Sprintf(" limit %d, %d ", (pageIndex-1)*pageSize, pageSize)
	if orderFiledMap == nil || len(orderFiledMap) == 0 {
		return limitSql
	}

	var orderSql []string
	for filedName, isAsc := range orderFiledMap {
		columnName := GetTableColumnNameByFieldName(modelType, filedName)
		if columnName == "" {
			continue
		}

		var sort = ""
		if isAsc {
			sort = " asc"
		} else {
			sort = " desc"
		}
		orderSql = append(orderSql, columnName+sort)
	}

	s := strings.Join(orderSql, ",")
	if len(s) > 0 {
		s = " order by " + s
	}

	return s + limitSql

}

func GetTableColumnNameByFieldName(modelType QueryModel, fieldName string) string {
	structType := reflect.TypeOf(modelType)
	columnName := ""
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if fieldName == field.Name {
			columnName = GetColumnName(field)
			break
		}
	}

	return columnName
}

func GetPageSize(paramMap map[string]interface{}) int64 {
	if utils.ContainKey(paramMap, "pageSize") {
		pageSize := utils.ToInt(fmt.Sprintf("%v", paramMap["pageSize"]), 10)
		if pageSize > 100000 {
			pageSize = 100000
		} else if pageSize < 1 {
			pageSize = 10
		}
		return pageSize
	} else {
		return 10
	}
}

func GetPageIndex(paramMap map[string]interface{}) int64 {
	var pageIndex int64
	var pageNum int64
	var pageNo int64
	if utils.ContainKey(paramMap, "pageIndex") {
		pageIndex = utils.ToInt(fmt.Sprintf("%v", paramMap["pageIndex"]), 1)
	}
	if utils.ContainKey(paramMap, "pageNum") {
		pageNum = utils.ToInt(fmt.Sprintf("%v", paramMap["pageNum"]), 1)
	}
	if utils.ContainKey(paramMap, "pageNo") {
		pageNo = utils.ToInt(fmt.Sprintf("%v", paramMap["pageNo"]), 1)
	}

	pageIndex = mathutil.MaxInt64(pageIndex, pageNum)
	pageIndex = mathutil.MaxInt64(pageIndex, pageNo)
	if pageIndex < 1 {
		pageIndex = 1
	}
	return pageIndex
}

func GetJsonProName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return field.Name
	}
	return tag
}

func GetColumnName(field reflect.StructField) string {
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

func GetPrimaryKeyJsonName(modelType QueryModel) string {
	structType := reflect.TypeOf(modelType)
	var fieldName string
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		columnName := field.Tag.Get("column")
		jsonName := field.Tag.Get("json")
		if strings.Contains(field.Tag.Get("gorm"), "primaryKey") {
			fieldName = field.Name
			if jsonName != "" {
				return jsonName
			} else if columnName != "" {
				return columnName
			} else {
				return GetColumnName(field)
			}
		}
	}
	return fieldName
}

func GetModelType(dest interface{}) reflect.Type {

	value := reflect.ValueOf(dest)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		value = reflect.New(value.Type().Elem())
	}
	modelType := reflect.Indirect(value).Type()

	if modelType.Kind() == reflect.Interface {
		modelType = reflect.Indirect(reflect.ValueOf(dest)).Elem().Type()
	}

	for modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		if modelType.PkgPath() == "" {
			return nil //, fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
		}
		return nil //, fmt.Errorf("%w: %s.%s", ErrUnsupportedDataType, modelType.PkgPath(), modelType.Name())
	}

	return modelType //, nil
}
