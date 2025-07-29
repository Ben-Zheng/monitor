package util

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMtimeduration(t *testing.T) {
	now := time.Now()
	currentTimestamp := now.Unix()
	fifteenDaysAgo := now.AddDate(0, 0, -30)
	fifteenDaysAgoTimestamp := fifteenDaysAgo.Unix()
	To := strconv.FormatInt(currentTimestamp, 10)
	From := strconv.FormatInt(fifteenDaysAgoTimestamp, 10)

	fmt.Println(To, From)
}

func TestReplaceModes(t *testing.T) {
	modelName := "qwen"
	var expr string
	expr = fmt.Sprintf("process_success{llm_model=~\"%s.*\"}", modelName)
	fmt.Println(expr)
}

func TestReplaceIKey(t *testing.T) {
	b := "node,modelName"
	a := "1233"
	fmt.Println(b + a)
}
func TestProscessStr(t *testing.T) {
	a := "(已下线)01jxm5r2#交银重点产品个性化方案推荐生成助手#Qwen2.5-vl-32B(910b)"
	b := ProcessSceneString(a)
	fmt.Println(b)
}

func TestGenerateAPIKey(t *testing.T) {
	dateStr := "Thu Jul 10 2025 00:00:00 GMT+0800 (中国标准时间)"

	// 步骤1：清理字符串，移除括号内容
	cleaned := strings.Split(dateStr, " (")[0]
	cleaned = strings.TrimSpace(cleaned)

	// 步骤2：定义解析格式（使用Go的布局模式）
	layout := "Mon Jan 02 2006 15:04:05 GMT-0700"

	// 步骤3：解析时间
	tt, err := time.Parse(layout, cleaned)
	if err != nil {
		panic(fmt.Errorf("解析失败: %v", err))
	}

	// 步骤4：转换为毫秒级时间戳
	timestamp := tt.UnixNano() / int64(time.Millisecond)
	fmt.Println(timestamp) // 输出: 1751702400000
}

func TestGetTime(t *testing.T) {
	dataStr := "Thu Jul 10 2025 00:00:00 GMT+0800 (中国标准时间)"
	exp, err := ParseTimeInput(dataStr, time.Now())
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(exp)
}
