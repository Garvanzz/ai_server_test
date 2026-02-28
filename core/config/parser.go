package config

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"reflect"
	"strconv"
	"strings"
	"xfx/pkg/log"
)

// ParseToStruct 解析到结构体
func ParseToStruct[T any](raw any, jsonName string) any {
	data := raw.([]byte)

	result := gjson.Parse(string(data))

	switch {
	case result.IsArray():
		return parseArrayToMap[T](result, jsonName)
	case result.IsObject():
		return parseSingleObject[T](result, jsonName)
	default:
		log.Fatal("parse to struct,data error")
		return nil
	}
}

func parseSingleObject[T any](result gjson.Result, jsonName string) T {
	var v T
	if err := json.Unmarshal([]byte(result.Raw), &v); err != nil {
		log.Fatal("json:%v, parseSingleObject: %v", jsonName, err)
	}
	return v
}

func parseArrayToMap[T any](result gjson.Result, jsonName string) map[int64]T {
	mapping := make(map[int64]T)

	result.ForEach(func(_, value gjson.Result) bool {
		idResults := gjson.GetMany(value.Raw, []string{"id", "iD", "Id", "ID"}...)

		ids := make([]int64, 0)
		for _, idResult := range idResults {
			if idResult.Exists() && idResult.Int() != 0 {
				ids = append(ids, idResult.Int())
			}
		}

		if len(ids) > 1 {
			log.Error("excel:%v, json:%v, to many id:%v", excelMap[jsonName], jsonName, ids)
			return false
		}
		if len(ids) == 0 {
			log.Error("excel:%v, json:%v,no id", excelMap[jsonName], jsonName)
			return false
		}

		id := ids[0]

		var item T
		if err := json.Unmarshal([]byte(value.Raw), &item); err != nil {
			log.Error("excel:%v, json:%v, parseArrayToMap json unmarshal error:%v", excelMap[jsonName], jsonName, err)
			return false
		}

		if _, exists := mapping[id]; exists {
			log.Error("excel:%v, json:%v, parseArrayToMap id repeated:%v", excelMap[jsonName], jsonName, id)
			return false
		}

		mapping[id] = item
		return true
	})

	return mapping
}

// AttachToStruct 绑定到结构体
func AttachToStruct[T any](raw any, jsonName string) T {
	data := raw.([]byte)
	if data[0] != '[' {
		log.Fatal("excel:%v, json:%v, attach to struct data is not list", excelMap[jsonName], jsonName)
	}

	reader := make([]map[string]any, 0)
	if err := json.Unmarshal(data, &reader); err != nil {
		log.Fatal("%v", err)
	}

	temp := make(map[string]any)
	for _, v := range reader {

		idStr, content, dataType := v["id"].(string), v["content"].(string), v["type"].(string)
		switch dataType {
		case "int":
			i, _ := strconv.ParseInt(content, 10, 64)
			temp[idStr] = i
		case "float":
			f, _ := strconv.ParseFloat(content, 64)
			temp[idStr] = f
		case "items":
			res := make([]map[string]any, 0)
			arr := strings.Split(content, ",")
			for _, item := range arr {
				kv := make(map[string]any)
				pair := strings.Split(item, "|")
				kv["itemId"], _ = strconv.ParseInt(pair[0], 10, 64)
				kv["itemNum"], _ = strconv.ParseInt(pair[1], 10, 64)
				kv["itemType"], _ = strconv.ParseInt(pair[2], 10, 64)
				res = append(res, kv)
			}
			temp[idStr] = res
		case "int[][]":
			arrList := strings.Split(content, ",")
			l := make([][]int32, 0)
			for _, arr := range arrList {
				ll := make([]int32, 0)
				for _, str := range strings.Split(arr, "|") {
					i, _ := strconv.Atoi(str)
					ll = append(ll, int32(i))
				}
				l = append(l, ll)
			}
			temp[idStr] = l
		case "int[]":
			l := make([]int32, 0)
			strList := strings.Split(content, "|")
			for _, str := range strList {
				s, _ := strconv.Atoi(str)
				l = append(l, int32(s))
			}
			temp[idStr] = l
		case "string":
			temp[idStr] = content
		default:
			log.Error("excel:%v, json:%v, type error:%v", excelMap[jsonName], jsonName, dataType)
		}
	}

	b, err := json.Marshal(temp)
	if err != nil {
		log.Fatal("excel:%v, json:%v,marshal error:%v", excelMap[jsonName], jsonName, err)
	}

	res := new(T)
	err = json.Unmarshal(b, res)
	if err != nil {
		log.Fatal("excel:%v, jsonName:%v,%v", excelMap[jsonName], jsonName, err)
	}

	return *res
}

func StructToStringMap(value interface{}) map[string]string {
	tempMap := make(map[string]string)
	attr := reflect.TypeOf(value)
	attrValue := reflect.ValueOf(value)
	for k := 0; k < attr.NumField(); k++ {
		tempMap[attr.Field(k).Name] = fmt.Sprint(attrValue.Field(k))
	}

	return tempMap
}
