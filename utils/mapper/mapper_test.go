package mapper

import (
	"encoding/json"
	"testing"
)

// ============================================================================
// v2 新特性测试: @ 语法与智能默认值
// ============================================================================

func TestDecode_SmartDefaults_V2(t *testing.T) {
	type Config struct {
		Port  int     `json:"port @8080"`
		Debug bool    `json:"@true"`
		Host  string  `json:"host @ip"`
		Retry int     `json:"retry @3"` // 逗号兼容
		Ratio float64 `json:"ratio @0.5"`
	}

	// Case 1: 纯默认值
	inputEmpty := map[string]interface{}{}
	var cfg1 Config
	if err := Decode(inputEmpty, &cfg1); err != nil {
		t.Fatalf("Failed to decode defaults: %v", err)
	}

	if cfg1.Port != 8080 {
		t.Errorf("Port failed. Got %d, want 8080", cfg1.Port)
	}
	if !cfg1.Debug {
		t.Errorf("Debug failed. Got %v, want true", cfg1.Debug)
	}
	if cfg1.Retry != 3 {
		t.Errorf("Retry failed. Got %d, want 3", cfg1.Retry)
	}
	if cfg1.Ratio != 0.5 {
		t.Errorf("Ratio failed. Got %f, want 0.5", cfg1.Ratio)
	}

	// Case 2: 引用默认值
	inputRef := map[string]interface{}{
		"ip": "192.168.1.100",
	}
	var cfg2 Config
	if err := Decode(inputRef, &cfg2); err != nil {
		t.Fatalf("Failed to decode refs: %v", err)
	}
	if cfg2.Host != "192.168.1.100" {
		t.Errorf("Host ref failed. Got %s, want 192.168.1.100", cfg2.Host)
	}
}

// 修复后的 Tag 解析测试: 确保解析器已更新
func TestDecode_TagSyntaxStyles(t *testing.T) {
	type StyleStruct struct {
		A int `json:"a @1"` // 之前报错的地方
		B int `json:"b @2"`
		C int `json:"c @3"`
		D int `json:"id @4"`
		E int `json:"@5"`
	}

	var res StyleStruct
	Decode(map[string]interface{}{}, &res)

	if res.A != 1 {
		t.Errorf("Space separator failed! Got %d want 1. (Update mapper.go parseTag)", res.A)
	}
	if res.B != 2 || res.C != 3 || res.D != 4 || res.E != 5 {
		t.Errorf("Tag styles failed. Got: %+v", res)
	}
}

// ============================================================================
// 核心功能回归测试
// ============================================================================

func TestDecode_RemainLogic(t *testing.T) {
	type User struct {
		Name   string                 `json:"name"`
		RealID string                 `json:"id @uuid"`
		Remain map[string]interface{} `json:"-"`
	}

	input := map[string]interface{}{
		"name":      "Alice",
		"uuid":      "user-123",
		"UserAge":   18,
		"meta_info": "foo",
	}

	var u User
	if err := Decode(input, &u); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if u.RealID != "user-123" {
		t.Errorf("Reference failed, got %s", u.RealID)
	}
	if _, ok := u.Remain["uuid"]; ok {
		t.Error("Referenced key 'uuid' should NOT be in Remain")
	}
	if val, ok := u.Remain["UserAge"]; !ok || val != 18 {
		t.Errorf("Dynamic key 'UserAge' missing. Remain: %v", u.Remain)
	}
}

func TestDecode_DeepNestingAndNumber(t *testing.T) {
	type NumStruct struct {
		Val int64 `json:"val"`
	}
	inputNum := map[string]interface{}{
		"val": json.Number("1234567890123"),
	}
	var resNum NumStruct
	Decode(inputNum, &resNum)
	if resNum.Val != 1234567890123 {
		t.Errorf("json.Number failed, got %v", resNum.Val)
	}
}
