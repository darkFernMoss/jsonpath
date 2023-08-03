package main

import (
	"encoding/json"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"sort"
	"strings"
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
	lines := findHitLines(string(dataJson), string(alarmJson))
	fmt.Println(lines)
}

func findHitLines(jsonData, alarmJson string) (ans []int) {
	logrus.SetReportCaller(true)
	var kvSlice []alarmKV
	err := json.Unmarshal([]byte(alarmJson), &kvSlice)
	if err != nil {
		logrus.WithError(err).Fatal("the format of alarmJson is wrong")
		return nil
	}
	dataMap := orderedmap.New()
	dataMap.UnmarshalJSON([]byte(jsonData))
	for _, kv := range kvSlice {
		keys := strings.Split(kv.Key, ".")
		ok, line := find(keys, kv.Value, dataMap, 0)
		if ok {
			flag := false
			for _, i := range ans {
				if i == line+1 {
					flag = true
					break
				}
			}
			if !flag {
				ans = append(ans, line+1)
			}
		}
	}
	sort.Ints(ans)
	return
}

func find(keys []string, value interface{}, dataMap *orderedmap.OrderedMap, curLine int) (ok bool, line int) {
	itf, ok := dataMap.Get(keys[0])
	if !ok {
		return false, curLine
	}
	mapKeys := dataMap.Keys()
	for _, k := range mapKeys {
		if k != keys[0] {
			get, _ := dataMap.Get(k)
			curLine += countLine(get)
		} else {
			break
		}
	}
	if len(keys) == 1 {
		return compareValues(itf, value), curLine + 1
	}
	sonMap, ok := itf.(orderedmap.OrderedMap)
	if ok {
		return find(keys[1:], value, &sonMap, curLine+1)
	}
	return false, curLine
}

func countLine(itf interface{}) (ans int) {
	omap, ok := itf.(orderedmap.OrderedMap)
	if ok {
		keys := omap.Keys()
		for _, k := range keys {
			get, _ := omap.Get(k)
			ans += countLine(get)
		}
		return ans + 2
	}
	arr, ok := itf.([]interface{})
	if ok {
		if len(arr) == 0 {
			return ans + 1
		}
		for _, obj := range arr {
			ans += countLine(obj)
		}
		return ans + 2
	}
	return 1
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

type alarmKV struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
