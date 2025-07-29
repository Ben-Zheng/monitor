package model

import (
	"context"
	"fmt"
	"monitor/config"
	"monitor/internal/client"
	"monitor/internal/types"
	"time"
)

type ModelQueryReq struct {
	ctx  context.Context
	from int64
	to   int64
}

func NewModelQueryReq(ctx context.Context, from int64, to int64) *ModelQueryReq {
	return &ModelQueryReq{
		ctx:  ctx,
		from: from,
		to:   to,
	}
}

func (m *ModelQueryReq) GetPodInfo(data_name, keyword string) (*types.VectorResponse, error) {
	query := config.NewQueryConfig()
	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)
	// 生成秒级Unix时间戳（10位字符串格式）
	nowTimestamp := fmt.Sprintf("%d", now.Unix())
	fiveMinutesAgoTimestamp := fmt.Sprintf("%d", fiveMinutesAgo.Unix())

	queryStr := BuildModelQueryRange(data_name, keyword, fiveMinutesAgoTimestamp, nowTimestamp)
	client := client.NewDCEClient(m.ctx, config.GetGrafanaQueryConfig().ClusterBaseURL, query.InsecureSkipVerify)
	result, err := client.MakeGetReqRange(queryStr, "/apis/insight.io/v1alpha1/metric/queryrange")
	if err != nil {
		fmt.Println(err)
	}
	return result, err
}
