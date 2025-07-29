package config

// 查询参数配置结构体
type QueryConfig struct {
	LegendFormat       string
	RefID              string
	DatasourceID       int
	DatasourceUID      string
	IntervalMs         int64
	MaxDataPoints      int64
	InsecureSkipVerify bool
}

func NewQueryConfig() *QueryConfig {
	gc := GetGrafanaQueryConfig()
	return &QueryConfig{
		DatasourceID:       gc.DatasourceID,
		IntervalMs:         gc.IntervalMs,
		MaxDataPoints:      gc.MaxDataPoints,
		RefID:              gc.RefID,
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
