package client

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"monitor/config"
	"monitor/internal/types"
	"strconv"
	"time"
)

type DCEClient struct {
	ctx    context.Context
	client *resty.Client
}

func NewDCEClient(ctx context.Context, dceURL string, skipVerify bool) *DCEClient {
	c := resty.New().
		SetRetryCount(1).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(3 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: skipVerify}).
		SetBaseURL(dceURL)
	return &DCEClient{client: c, ctx: ctx}
}

func getDceToken(ctx context.Context) string {
	return config.GetGrafanaQueryConfig().Token
}

func (c *DCEClient) MakeGetReqRange(query map[string]string, url string) (*types.VectorResponse, error) {
	resp, err := c.client.R().SetAuthToken(getDceToken(c.ctx)).
		SetQueryParam("query", query["query"]).
		SetQueryParam("start", query["start"]).
		SetQueryParam("end", query["end"]).
		SetQueryParam("step", strconv.Itoa(60)).Get(url)
	if err != nil {
		log.Println(err)
	}

	respBody := resp.Body()
	var decompressed []byte
	if len(respBody) > 2 && respBody[0] == 0x1f && respBody[1] == 0x8b {
		// GZIP 魔术头匹配 → 解压
		gr, err := gzip.NewReader(bytes.NewReader(respBody))
		if err != nil {
			log.Fatal(err)
		}
		defer gr.Close()
		decompressed, err = ioutil.ReadAll(gr)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		decompressed = respBody
	}
	result := &types.VectorResponse{}
	err = json.Unmarshal(decompressed, result)
	// 使用辅助函数填充缺失字段
	return result, err
}

func (c *DCEClient) GetClusterName(urlStr string) []types.NameList {
	f := time.Now().Unix()
	t := time.Now().Add(-5 * time.Minute).Unix()
	nowStr := strconv.FormatInt(f, 10)
	fiveMinutesAgoStr := strconv.FormatInt(t, 10)

	query := map[string]string{
		"query": "kpanda_gpu_count",
		"start": fiveMinutesAgoStr,
		"end":   nowStr,
	}

	resp, err := c.client.R().SetAuthToken(getDceToken(c.ctx)).
		SetQueryParam("query", query["query"]).
		SetQueryParam("start", query["start"]).
		SetQueryParam("end", query["end"]).
		SetQueryParam("step", strconv.Itoa(60)).Get(urlStr)

	if err != nil {
		log.Println(err)
	}
	respBody := resp.Body()
	var decompressed []byte
	if len(respBody) > 2 && respBody[0] == 0x1f && respBody[1] == 0x8b {
		// GZIP 魔术头匹配 → 解压
		gr, err := gzip.NewReader(bytes.NewReader(respBody))
		if err != nil {
			log.Fatal(err)
		}
		defer gr.Close()
		decompressed, err = ioutil.ReadAll(gr)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		decompressed = respBody
	}
	var result types.VectorResponse
	err = json.Unmarshal(decompressed, &result)
	nMap := make(map[string]string)
	for i := range result.Matrix {
		clustername := result.Matrix[i].Metric.ClusterName + result.Matrix[i].Metric.Cluster_Name
		cluster := result.Matrix[i].Metric.Cluster
		nMap[cluster] = clustername
	}

	nameList := make([]types.NameList, 0)
	for k, v := range nMap {
		nameList = append(nameList, types.NameList{
			Cluster:     k,
			ClusterName: v,
		})
	}

	return nameList
}

func (c *DCEClient) MakePostReqRange(query map[string]string, url string) (types.VectorResponse, error) {
	resp, err := c.client.R().SetAuthToken(getDceToken(c.ctx)).
		SetQueryParam("query", query["query"]).
		SetQueryParam("start", query["start"]).
		SetQueryParam("end", query["end"]).
		SetQueryParam("step", strconv.Itoa(60)).Post(url)
	if err != nil {
		log.Println(err)
	}

	respBody := resp.Body()
	var decompressed []byte
	if len(respBody) > 2 && respBody[0] == 0x1f && respBody[1] == 0x8b {
		// GZIP 魔术头匹配 → 解压
		gr, err := gzip.NewReader(bytes.NewReader(respBody))
		if err != nil {
			log.Fatal(err)
		}
		defer gr.Close()
		decompressed, err = ioutil.ReadAll(gr)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		decompressed = respBody
	}
	var result types.VectorResponse
	err = json.Unmarshal(decompressed, &result)
	// 使用辅助函数填充缺失字段
	return result, err
}
