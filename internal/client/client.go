package client

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
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

// 定义接口响应结构
type ApiResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    ResponseData `json:"result"`
}

type ResponseData struct {
	Current int         `json:"current"` // 当前页码
	Size    int         `json:"size"`    // 每页数量
	Pages   int         `json:"pages"`
	Total   int         `json:"total"`   // 总数据量
	Items   []TokenItem `json:"records"` // 实际数据项
}

// 定义接口请求结构
type pageparam struct {
	page     int `json:"pageNum"`  // 当前页码
	pageSize int `json:"pageSize"` // 每页数量
}

type orderparam struct {
	column string `json:"column"`
	order  string `json:"order"`
}

type ReqParam struct {
	PageParam  pageparam  `json:"pageParam"`
	OrderParam orderparam `json:"orderParam"`
}

type TokenItem struct {
	Id                 string `json:"id"`
	CallModelName      string `json:"callModelName"`
	ApisixScenarioName string `json:"apisixScenarioName"`
	CallModelId        string `json:"callModelId"`
	DevDept            string `json:"devDept"`
	DevManager         string `json:"devManager"`
	EnvAlias           string `json:"envAlias"`
	EnvName            string `json:"envName"`
	ModelName          string `json:"modelName"`
	Token              string `json:"token"`
	MaxConcurrency     int    `json:"maxConcurrency"`
	Status             string `json:"status"`
}

// 获取场景管理分页接口的所有数据
func (c *DCEClient) GetSceneManageInfo(url string, pageSize int) ([]TokenItem, error) {
	// 初始化返回结果
	var allItems []TokenItem

	// 设置起始页码
	currentPage := 1
	totalPages := 1
	var reqParam ReqParam
	for currentPage <= totalPages {
		// 准备请求体
		reqParam.PageParam = pageparam{currentPage, pageSize}
		reqParam.OrderParam = orderparam{"createTime", "decs"}

		resp, err := c.client.R().SetAuthToken(getDceToken(c.ctx)).
			SetBody(reqParam).Post(url)
		if err != nil {
			log.Println(err)
		}

		// 读取响应体
		bodyBytes := resp.Body()
		if err != nil {
			return nil, fmt.Errorf("读取响应体失败: %v", err)
		}

		// 解析响应
		var response ApiResponse
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
			return nil, fmt.Errorf("解析JSON失败: %v, 原始响应: %s", err, string(bodyBytes))
		}
		// 首次请求时设置总页数
		if currentPage == 1 {
			totalItems := response.Data.Total
			totalPages = response.Data.Pages
			// 预分配切片容量以优化性能
			allItems = make([]TokenItem, 0, totalItems)
		}

		// 检查API响应状态码
		if response.Code != 0 && response.Code != 200 {
			return nil, fmt.Errorf("接口返回错误: 代码: %d, 消息: %s", response.Code, response.Message)
		}

		// 添加到总结果
		allItems = append(allItems, response.Data.Items...)

		// 更新页码
		currentPage++
	}

	return allItems, nil
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
