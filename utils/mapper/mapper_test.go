package mapper

import (
	"encoding/json"
	"testing"
	"time"
)

// 测试 Map Key 的类型转换 (字符串 -> Int)
func TestDecode_MapKeyConversion(t *testing.T) {
	type Config struct {
		// JSON 中 key 总是 string, 但 struct 中我们想要 int
		Weights map[int]string `json:"weights"`
	}

	input := map[string]interface{}{
		"weights": map[string]interface{}{
			"1":  "Low",
			"10": "High",
		},
	}

	var result Config
	if err := Decode(input, &result); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if result.Weights[1] != "Low" || result.Weights[10] != "High" {
		t.Errorf("Map key conversion failed. Got: %v", result.Weights)
	}
}

// 测试灵活的时间格式
func TestDecode_FlexibleTime(t *testing.T) {
	type Event struct {
		T1 time.Time `json:"t1"`
		T2 time.Time `json:"t2"`
	}

	input := map[string]interface{}{
		"t1": "2023-12-01",          // 纯日期
		"t2": "2023-12-01 10:00:00", // 空格分隔
	}

	var result Event
	if err := Decode(input, &result); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if result.T1.Year() != 2023 {
		t.Error("Failed to parse simple date")
	}
}
func TestDecode_DeepNestingAndNumber(t *testing.T) {
	// 1. 测试 json.Number
	type NumStruct struct {
		Val int64 `json:"val"`
	}
	// 模拟 UseNumber() 产生的输入
	inputNum := map[string]interface{}{
		"val": json.Number("1234567890123"),
	}
	var resNum NumStruct
	Decode(inputNum, &resNum)
	if resNum.Val != 1234567890123 {
		t.Errorf("json.Number failed, got %v", resNum.Val)
	}

	// 2. 测试深层匿名结构体 (Deep Squash)
	type Base struct {
		ID string `json:"id"`
	}
	type Middle struct {
		Base
		Tag string `json:"tag"`
	}
	type Top struct {
		Middle
		Name string `json:"name"`
	}

	inputSquash := map[string]interface{}{
		"id":   "001",
		"tag":  "algo",
		"name": "doptime",
	}
	var resSquash Top
	Decode(inputSquash, &resSquash)

	if resSquash.ID != "001" || resSquash.Tag != "algo" {
		t.Errorf("Deep squash failed: %+v", resSquash)
	}
}
