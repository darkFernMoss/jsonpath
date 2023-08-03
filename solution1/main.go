package main

import (
	"bytes"
	"darkFernMoss/jsonHighLight/solution1/ordermap"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"strings"
)

// [3 4 6 7 8 10 11 12 15 16 17 195 235]
func main() {
	dataJson, err := os.ReadFile("jsonHighLight/data.json")
	if err != nil {
		logrus.WithError(err).Fatalln()
	}
	alarmJson, err := os.ReadFile("jsonHighLight/alarm.json")
	if err != nil {
		logrus.WithError(err).Fatalln()
	}
	_, lines, _ := findHitLines(string(dataJson), string(alarmJson))
	fmt.Println(lines)
}

func findHitLines(sourceJson, alarmJson string) (formatJson string, lines []int, err error) {
	result, err := processJSON(alarmJson, sourceJson)
	if err != nil {
		return "", nil, err
	}
	bufLines := bytes.Buffer{}
	err = json.Indent(&bufLines, []byte(result), "", "    ")
	if err != nil {
		return "", nil, err
	}
	formatJs, lines := getLinesAndFormatJson(bufLines.String())
	return formatJs, lines, nil
}

// 把目标key替换为目标key_cspm_highlight
func processJSON(alarmJson, jsonData string) (string, error) {
	m := ordermap.New()
	err := json.Unmarshal([]byte(jsonData), &m)
	if err != nil {
		return "", err
	}

	kvSlice := []alarmKV{}
	err = json.Unmarshal([]byte(alarmJson), &kvSlice)
	if err != nil {
		return "", err
	}

	for _, kv := range kvSlice {
		splitK := strings.Split(kv.Key, ".")
		tmpM := m
		for i, k := range splitK {
			if val, ok := tmpM.Get(k); ok {
				if i == len(splitK)-1 {
					isMatch := compareValues(val, kv.Value)
					if isMatch {
						newK := fmt.Sprintf("%s_cspm_highlight", k)
						tmpM.ReplaceKey(k, newK)
					}
				} else {
					if valMap, ok := val.(ordermap.OrderedMap); ok {
						tmpM = &valMap
					} else if sliceMap, ok := val.([]interface{}); ok {
						// 如果不是map, 需要直接处理, 处理完把整个value替换
						var newSliceMap []ordermap.OrderedMap
						for _, v := range sliceMap {
							tmpV, ok := v.(ordermap.OrderedMap)
							if !ok {
								continue
							}
							tmpK := splitK[i+1]
							if val, ok := tmpV.Get(tmpK); ok {
								isMatch := compareValues(val, kv.Value)
								if isMatch {
									newK := fmt.Sprintf("%s_cspm_highlight", tmpK)
									tmpV.ReplaceKey(tmpK, newK)
								}
							}
							newSliceMap = append(newSliceMap, tmpV)

						}
						tmpM.Set(k, newSliceMap)
						break
					}
				}
			} else {
				break
			}
		}
	}

	result, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func getLinesAndFormatJson(src string) (string, []int) {
	splits := strings.Split(src, "\n")
	result := []int{}
	for i, line := range splits {
		if strings.Contains(line, "_cspm_highlight") {
			result = append(result, i+1)
		}
	}
	format := strings.ReplaceAll(src, "_cspm_highlight", "")
	return format, result
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
