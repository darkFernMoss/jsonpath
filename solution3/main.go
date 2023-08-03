package main

import (
	"encoding/json"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
)

func main() {
	dataJson, err := os.ReadFile("jsonHighLight/data.json")
	if err != nil {
		logrus.WithError(err).Errorln()
	}
	alarmJson, err := os.ReadFile("jsonHighLight/alarm.json")
	if err != nil {
		logrus.WithError(err).Errorln()
	}
	_, lines, _ := findHitLines(string(dataJson), string(alarmJson))
	for _, v := range lines {
		fmt.Printf(" %d", v)
	}
}

func findHitLines(rawData, alarmData string) (string, map[string]int, error) {
	var kvs []alarmKV
	err := json.Unmarshal([]byte(alarmData), &kvs)
	if err != nil {
		return "", nil, err
	}

	// 如果 Key 会重复，则构造 map[string][]interface{}
	alarms := make(map[string]interface{})
	for _, kv := range kvs {
		alarms[kv.Key] = kv.Value
	}

	var dataMap orderedmap.OrderedMap
	err = json.Unmarshal([]byte(rawData), &dataMap)
	if err != nil {
		return "", nil, err
	}

	line := 1
	results := make(map[string]int)
	matchByFullPath(dataMap, "", &line, alarms, results)

	data, err := json.MarshalIndent(&dataMap, "", "    ")
	if err != nil {
		return "", nil, err
	}
	return string(data), results, nil
}

func matchByFullPath(data interface{}, path string, line *int, alarms map[string]interface{}, results map[string]int) {
	v, ok := alarms[path]
	if ok {
		if compareValues(data, v) {
			results[path] = *line
		}
	}

	dataMap, ok := data.(orderedmap.OrderedMap)
	if ok {
		// TODO: 确认 orderedmap.OrderedMap 的解析是否将空对象的 ']' 换行
		*line += b2i(len(dataMap.Keys()) != 0)
		for _, k := range dataMap.Keys() {
			v, _ := dataMap.Get(k)
			matchByFullPath(v, buildPath(path, k), line, alarms, results)
		}
		//*line += b2i(len(dataMap.Keys()) != 0)
		//return
	}

	l, ok := data.([]interface{})
	if ok {
		*line += b2i(len(l) != 0)
		for _, e := range l {
			matchByFullPath(e, path, line, alarms, results)
		}
		//*line += b2i(len(l) != 0)
		//return
	}
	*line++
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func buildPath(dir, file string) string {
	if len(dir) == 0 {
		return file
	}
	return fmt.Sprintf("%s.%s", dir, file)
}

type alarmKV struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func compareValues(a, b interface{}) bool {
	typeA := reflect.TypeOf(a)
	typeB := reflect.TypeOf(b)

	if typeA != typeB {
		return false
	}

	switch a.(type) {
	case string:
		return a.(string) == b.(string)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return reflect.DeepEqual(a, b)
	case []interface{}:
		sliceA := a.([]interface{})
		sliceB := b.([]interface{})
		if len(sliceA) == 0 && len(sliceB) == 0 {
			return true
		}
		// 有一个交叉及视为匹配
		for _, valA := range sliceA {
			for _, valB := range sliceB {
				if compareValues(valA, valB) {
					return true
				}
			}
		}
		return false

	case bool:
		return a.(bool) == b.(bool)
	default:
		return false
	}
}
