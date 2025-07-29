package util

import (
	"encoding/json"
	"fmt"
	"math"
	"monitor/internal/types"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Percentiles(data []float64, ps ...float64) (map[float64]float64, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("空切片：无法计算百分位数")
	}

	// 创建数据副本以避免修改原始切片
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	results := make(map[float64]float64, len(ps))

	for _, p := range ps {
		if p < 0 || p > 100 {
			return nil, fmt.Errorf("无效的百分位数: %.2f (必须在0-100范围内)", p)
		}

		if p == 0 {
			results[p] = sorted[0]
			continue
		}

		if p == 100 {
			results[p] = sorted[len(sorted)-1]
			continue
		}

		// 计算位置和插值
		position := (float64(len(sorted)) - 1) * p / 100
		lowerIndex := int(math.Floor(position))
		upperIndex := int(math.Ceil(position))

		if lowerIndex == upperIndex {
			results[p] = sorted[lowerIndex]
		} else {
			// 线性插值公式：value = sorted[lowerIndex] + fraction*(sorted[upperIndex]-sorted[lowerIndex])
			fraction := position - float64(lowerIndex)
			results[p] = sorted[lowerIndex] + fraction*(sorted[upperIndex]-sorted[lowerIndex])
		}
	}

	return results, nil
}
func RoundFloat64(val float64) float64 {
	return math.Round(val*100) / 100
}

func CalculateP50P90P95(data []float64) (p50, p90, p95 float64, err error) {
	percentilesMap, err := Percentiles(data, 50, 90, 95)
	if err != nil {
		return 0, 0, 0, err
	}
	return RoundFloat64(percentilesMap[50]), RoundFloat64(percentilesMap[90]), RoundFloat64(percentilesMap[95]), nil
}
func CalculateMedian(nums []int) (int, error) {
	// 处理空切片情况
	if len(nums) == 0 {
		return 0, fmt.Errorf("切片为空")
	}

	sort.Ints(nums)

	n := len(nums)

	return nums[n/2], nil
}

func CalculateAverage(nums []int) (int, error) {
	if len(nums) == 0 {
		return 0, fmt.Errorf("切片为空")
	}

	total := 0
	for _, num := range nums {
		total += num
	}

	// 使用向上取整的除法公式
	avg := int(math.Ceil(float64(total) / float64(len(nums))))
	return avg, nil
}

func SafePercent(part, total int) int {
	if total == 0 {
		// 根据业务需求返回0、NaN或特殊值
		return 0
	}
	return int(part) / int(total) * 100
}

// ParseRelativeTime 解析相对时间字符串为绝对时间
func ParseRelativeTime(input string, base time.Time) (time.Time, error) {
	// 支持now
	if input == "now" {
		return base, nil
	}

	// 支持now-语法
	re := regexp.MustCompile(`^now-(\d+)([smhdwMy])$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) != 3 {
		return time.Time{}, fmt.Errorf("无效的时间格式: %s", input)
	}

	value, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	// 根据单位计算时间偏移
	switch unit {
	case "s": // 秒
		return base.Add(time.Duration(-value) * time.Second), nil
	case "m": // 分钟
		return base.Add(time.Duration(-value) * time.Minute), nil
	case "h": // 小时
		return base.Add(time.Duration(-value) * time.Hour), nil
	case "d": // 天
		return base.AddDate(0, 0, -value), nil
	case "w": // 周
		return base.AddDate(0, 0, -7*value), nil
	case "M": // 月
		return base.AddDate(0, -value, 0), nil
	case "y": // 年
		return base.AddDate(-value, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("未知的时间单位: %s", unit)
	}
}

// ToMilliseconds 将时间转换为毫秒时间戳
func ToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// ParseTimeInput 处理各种时间输入格式
func ParseTimeInput(input string, base time.Time) (int64, error) {
	// 如果是相对时间格式
	if input == "now" || regexp.MustCompile(`^now-`).MatchString(input) {
		t, err := ParseRelativeTime(input, base)
		if err != nil {
			return 0, err
		}
		return ToMilliseconds(t), nil
	}

	// 尝试解析为RFC3339格式
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return ToMilliseconds(t), nil
	}

	// 尝试解析为时间戳
	if val, err := strconv.ParseInt(input, 10, 64); err == nil {
		// 确定时间戳的精度（秒、毫秒、微秒、纳秒）
		switch len(input) {
		case 10: // 10位秒级时间戳
			return val * 1000, nil
		case 13: // 13位毫秒级时间戳
			return val, nil
		case 16: // 16位微秒级时间戳
			return val / 1000, nil
		case 19: // 19位纳秒级时间戳
			return val / 1000000, nil
		}
	}

	// 尝试其他常见格式
	formats := []string{
		"2006-01-02T15:04:05.000Z", // ISO格式
		"2006-01-02 15:04:05",      // 无时区格式
		"2006-01-02",               // 日期格式
		"15:04:05",                 // 时间格式
		time.RFC1123,               // HTTP日期格式
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return ToMilliseconds(t), nil
		}
	}

	if strings.Contains(input, "GMT") && strings.Contains(input, "(") {
		cleaned := strings.Split(input, " (")[0]
		if t, err := time.Parse("Mon Jan 02 2006 15:04:05 GMT-0700", cleaned); err == nil {
			return ToMilliseconds(t), nil
		}
	}

	return 0, fmt.Errorf("无法解析的时间格式: %s", input)
}

// 总和在单独计算百分比
// 英伟达sum(DCGM_FI_DEV_FB_TOTAL)by(modelName)
// 昇腾floor(sum(npu_chip_info_hbm_total_memory)by(modelName)/1024
func MergeAndSumMaps(map1, map2 map[string]int) map[string]int {
	// 创建结果map
	result := make(map[string]int)

	// 复制第一个map的所有内容
	for key, value := range map1 {
		result[key] = value
	}

	// 遍历第二个map，并合并到结果map中
	for key, value := range map2 {
		// 如果键已存在，则值相加
		if existing, ok := result[key]; ok {
			result[key] = existing + value
		} else {
			// 如果键不存在，则直接添加
			result[key] = value
		}
	}

	return result
}

func ExtractValues(points []types.DataPoint) []int {
	var values []int

	for _, dp := range points {
		if val, err := strconv.Atoi(dp.Value); err == nil {
			values = append(values, val)
		}
	}

	return values
}

func ToString(val interface{}, defaultValue string) string {
	if val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case string:
		return v
	case json.Number: // 处理ES返回的数字类型
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return defaultValue
	}
}

func ToFloat64(val interface{}, defaultValue float64) float64 {
	if val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case float64:
		return v
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f
		}
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}

	return defaultValue
}

func ToInt64(str string) int64 {
	// 将字符串转换为 int64
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return 0
	}
	return num
}

func ProcessSceneString(input string) string {
	// 检查并保留开头的 "(已下线)"
	prefix := ""
	if strings.HasPrefix(input, "(已下线)") {
		prefix = "(已下线)"
		input = strings.TrimPrefix(input, "(已下线)")
	}

	// 找到第一个 # 的位置
	firstHash := strings.Index(input, "#")
	if firstHash == -1 {
		// 如果没有找到 #，返回前缀（如果有）
		return prefix
	}

	// 找到第二个 # 的位置
	secondHash := strings.Index(input[firstHash+1:], "#")
	if secondHash == -1 {
		// 如果只有一个 #，返回前缀和第一个 # 之后的内容
		return prefix + input[firstHash+1:]
	}

	// 调整第二个 # 的索引（因为是在子串中查找的）
	secondHash = firstHash + 1 + secondHash

	// 提取两个 # 之间的内容
	content := input[firstHash+1 : secondHash]

	return prefix + content
}

// SortDataPoints 按照时间戳排序 DataPoint 切片，返回最新的 value
func SortDataPoints(data []types.DataPoint) (string, error) {
	// 使用 sort.Slice 简单排序
	sort.Slice(data, func(i, j int) bool {
		// 将 Timestamp 字符串转换为 Unix 时间戳并比较
		t1, err1 := strconv.ParseInt(data[i].Timestamp, 10, 64)
		t2, err2 := strconv.ParseInt(data[j].Timestamp, 10, 64)
		if err1 != nil || err2 != nil {
			// 如果转换失败，选择适当的错误处理
			return false
		}
		// 使用 time.Unix 将 Unix 时间戳转换为 time.Time 类型进行比较
		return time.Unix(t1, 0).Before(time.Unix(t2, 0))
	})

	// 获取排序后的最新的 value
	if len(data) > 0 {
		return data[len(data)-1].Value, nil
	}

	return "", fmt.Errorf("data is empty")
}

func MergeMaps(m1, m2 map[string]int) map[string]int {
	result := make(map[string]int)

	// 添加第一个map的所有键值对
	for k, v := range m1 {
		result[k] = v
	}

	// 合并第二个map
	for k, v := range m2 {
		// 如果键已存在则求和，否则直接添加
		result[k] += v
	}
	return result
}
