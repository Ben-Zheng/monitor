package config

// 查询参数配置结构体
type QueryConfig struct {
	InsecureSkipVerify bool
}

func NewQueryConfig() *QueryConfig {
	gc := GetGrafanaQueryConfig()
	return &QueryConfig{
		InsecureSkipVerify: gc.InsecureSkipVerify,
	}
}

func (q *QueryConfig) BuildGrafanaQueryRange(expr, start, end string) map[string]string {
	return map[string]string{
		"query": expr,
		"start": start,
		"end":   end,
	}
}
