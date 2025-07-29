package es

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"log"
	"monitor/config"
	"monitor/internal/client"
	"time"
)

type EsRepo interface {
	Count(from, to int64, statusType string, reqType string, keyword string) (map[string]map[string]int64, error)
	CountByNestedAggs(from, to int64) (map[string]map[string]int64, error)
	GetDocumentFields(from, to int64, statusType string, sceneValue string, modelValue string) ([]map[string]interface{}, error)
	CountByModel(from, to int64) (map[string]map[string]int64, error)
}

type ESService struct {
	ESClient *client.ESClient
	Index    string
}

func NewESService(appConfig config.ESConfig) EsRepo {
	client, err := client.NewEsClient(appConfig)
	if err != nil {
		panic("es connect error")
	}
	return &ESService{
		ESClient: client,
		Index:    appConfig.Index,
	}
}

func (e *ESService) Count(from, to int64, statusType string, reqType string, keyword string) (map[string]map[string]int64, error) {
	if from == 0 || to == 0 || to <= from {
		return nil, fmt.Errorf("invalid time range: from=%d to=%d", from, to)
	}

	boolQuery := elastic.NewBoolQuery()
	boolQuery.Filter(elastic.NewRangeQuery("@timestamp").
		Gte(from).
		Lte(to).
		Format("epoch_millis"))

	switch statusType {
	case "success":
		boolQuery.Must(elastic.NewTermQuery("status", 200))
	case "failed":
		boolQuery.MustNot(elastic.NewTermQuery("status", 200))
	default:
	}

	boolQuery.MustNot(elastic.NewTermQuery("http_model.keyword", "-"))
	boolQuery.MustNot(elastic.NewTermQuery("http_model.keyword", ""))

	authAgg := elastic.NewTermsAggregation().
		Field("http_authorization.keyword").
		Size(500).
		OrderByKeyDesc().
		MinDocCount(1)

	modelAgg := elastic.NewTermsAggregation().
		Field("http_model.keyword").
		Size(500).
		OrderByKeyDesc().
		MinDocCount(1).
		Missing("N/A")

	var searchService *elastic.SearchService
	if reqType == "model" {
		boolQuery.Must(elastic.NewTermQuery("http_model.keyword", keyword))
		boolQuery.Must(elastic.NewTermQuery("method.keyword", "POST"))
		modelAgg = modelAgg.SubAggregation("http_info", authAgg)

		searchService = e.ESClient.Client.Search().
			Index(e.Index).
			Query(boolQuery).
			Size(0).
			IgnoreUnavailable(true).
			Aggregation("model_counts", modelAgg)

	} else if reqType == "scene" {
		boolQuery.Must(elastic.NewTermQuery("http_authorization.keyword", keyword))
		boolQuery.Must(elastic.NewTermQuery("method.keyword", "POST"))
		authAgg = authAgg.SubAggregation("http_info", modelAgg)

		searchService = e.ESClient.Client.Search().
			Index(e.Index).
			Query(boolQuery).
			Size(0).
			IgnoreUnavailable(true).
			Aggregation("model_counts", authAgg)
	}

	searchResult, err := searchService.Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ES query failed: %w", err)
	}

	modelRequestCountMap := make(map[string]map[string]int64)
	authTerms, found := searchResult.Aggregations.Terms("model_counts")
	if !found {
		return modelRequestCountMap, fmt.Errorf("model_counts")
	}

	for _, sceneName := range authTerms.Buckets {
		sceneKey, ok := sceneName.Key.(string)
		if !ok {
			sceneKey = fmt.Sprintf("%v", sceneName.Key) // 处理非字符串键
		}

		modelTerms, ok := sceneName.Aggregations.Terms("http_info")
		if !ok {
			log.Printf("授权分组 %s 中没有找到HTTP模型聚合", sceneKey)
			continue
		}
		modelMap := make(map[string]int64)

		for _, modelBucket := range modelTerms.Buckets {
			modelKey, ok := modelBucket.Key.(string)
			if !ok {
				modelKey = fmt.Sprintf("%v", modelBucket.Key)
			}
			modelMap[modelKey] = modelBucket.DocCount
		}
		modelRequestCountMap[sceneKey] = modelMap
	}

	return modelRequestCountMap, nil
}

// 1、所有场景下的，每个场景的绑定模型服务，本月调用次数，qps/时
// 2、单个场景下的，每个模型的本月调用次数，qps，报错日志
// 3、单个场景下的模型返回时间。
// map[scene][qwen:25,QWQ:20]
func (e *ESService) CountByNestedAggs(from, to int64) (map[string]map[string]int64, error) {
	var (
		query              = elastic.NewBoolQuery()
		termQueries        []elastic.Query
		mustNotTermQueries []elastic.Query
		methodKeyword      = "method.keyword"
		httpModelKeyword   = "http_model.keyword"
		authKeyword        = "http_authorization.keyword"
	)

	// 验证时间范围
	if from == 0 || to == 0 || to <= from {
		return nil, fmt.Errorf("invalid time range: from=%d to=%d", from, to)
	}

	termQueries = append(termQueries,
		elastic.NewTermsQuery(methodKeyword, "POST"))

	query = query.Must(termQueries...)
	query = query.Must(elastic.NewRangeQuery("@timestamp").From(from).To(to).Format("epoch_millis"))

	mustNotTermQueries = append(mustNotTermQueries,
		elastic.NewTermsQuery(httpModelKeyword, "-"))
	mustNotTermQueries = append(mustNotTermQueries,
		elastic.NewTermsQuery(httpModelKeyword, ""))
	query = query.MustNot(mustNotTermQueries...)

	src, err := query.Source()
	if err != nil {
		log.Printf("构建查询失败: %v", err)
	} else {
		jsonStr, _ := json.MarshalIndent(src, "", "  ")
		fmt.Println("基本查询 DSL:", string(jsonStr))
	}

	authAgg := elastic.NewTermsAggregation().
		Field(authKeyword).
		Size(500).
		OrderByKeyDesc().
		MinDocCount(1)

	modelAgg := elastic.NewTermsAggregation().
		Field(httpModelKeyword).
		Size(500).
		OrderByKeyDesc().
		MinDocCount(1)

	authAgg = authAgg.SubAggregation("http_models", modelAgg)
	Str, _ := authAgg.Source()
	log.Println(Str)
	aggsName := "nested_authorization_models"
	searchService := e.ESClient.Client.Search().
		Index(e.Index).
		Query(query).
		Size(0).
		IgnoreUnavailable(true).
		SearchType("query_then_fetch").
		Aggregation(aggsName, authAgg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("ES查询失败: %v", err)
	}
	modelRequestCountMap := make(map[string]map[string]int64)

	authTerms, found := searchResult.Aggregations.Terms(aggsName)
	if !found {
		return modelRequestCountMap, fmt.Errorf("未找到聚合结果")
	}

	for _, sceneName := range authTerms.Buckets {
		sceneKey, ok := sceneName.Key.(string)
		if !ok {
			sceneKey = fmt.Sprintf("%v", sceneName.Key) // 处理非字符串键
		}

		modelTerms, ok := sceneName.Aggregations.Terms("http_models")
		if !ok {
			log.Printf("授权分组 %s 中没有找到HTTP模型聚合", sceneKey)
			continue
		}
		modelMap := make(map[string]int64)
		for _, modelBucket := range modelTerms.Buckets {
			modelKey, ok := modelBucket.Key.(string)
			if !ok {
				modelKey = fmt.Sprintf("%v", modelBucket.Key)
			}
			modelMap[modelKey] = modelBucket.DocCount
		}
		modelRequestCountMap[sceneKey] = modelMap
	}

	return modelRequestCountMap, nil
}

func (e *ESService) GetDocumentFields(from, to int64, statusType string, sceneValue string, modelValue string) ([]map[string]interface{}, error) {
	query := elastic.NewBoolQuery()

	rangeQuery := elastic.NewRangeQuery("@timestamp").
		Format("epoch_millis").
		Gte(from).
		Lte(to)
	query = query.Filter(rangeQuery)

	if modelValue != "" && sceneValue != "" {
		query = query.Must(
			elastic.NewTermQuery("method.keyword", "POST"),
			elastic.NewTermQuery("http_authorization.keyword", sceneValue),
			elastic.NewTermQuery("http_model.keyword", modelValue),
		)
	} else if sceneValue != "" {
		query = query.Must(
			elastic.NewTermQuery("method.keyword", "POST"),
			elastic.NewTermQuery("http_authorization.keyword", sceneValue),
		)
	} else if modelValue != "" {
		query = query.Must(
			elastic.NewTermQuery("method.keyword", "POST"),
			elastic.NewTermQuery("http_model.keyword", modelValue),
		)
	} else {
		query = query.Must(
			elastic.NewTermQuery("method.keyword", "POST"),
		)
	}
	if statusType != "all" {
		switch statusType {
		case "success":
			query = query.Must(elastic.NewTermQuery("status", "200"))
		case "failed":
			query = query.MustNot(elastic.NewTermQuery("status", "200"))
		}
	}

	query = query.MustNot(elastic.NewTermQuery("http_model.keyword", "-"))
	query = query.MustNot(elastic.NewTermQuery("http_model.keyword", ""))

	// 指定要返回的字段
	fields := []string{"_id", "request_time", "upstream_response_time", "request_uri", "path", "time", "status", "http_model", "http_host", "request", "http_authorization"}

	searchService := e.ESClient.Client.Search().
		Index(e.Index).
		Query(query).
		Size(100).
		IgnoreUnavailable(true).
		FetchSourceContext(elastic.NewFetchSourceContext(true).Include(fields...)).
		Sort("@timestamp", false)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("ES查询失败: %v", err)
	}

	results := make([]map[string]interface{}, 0)
	if searchResult.Hits == nil || searchResult.Hits.TotalHits.Value == 0 {
		return results, nil
	}

	for _, hit := range searchResult.Hits.Hits {
		result := make(map[string]interface{})
		if err := json.Unmarshal(hit.Source, &result); err != nil {
			log.Printf("解析文档失败: %v", err)
			continue
		}
		for _, field := range fields {
			if value, exists := result[field]; exists {
				result[field] = value
			} else {
				result[field] = nil
			}
		}
		result["_id"] = hit.Id
		results = append(results, result)

	}
	return results, nil
}

func (e *ESService) CountByModel(from, to int64) (map[string]map[string]int64, error) {
	var (
		query              = elastic.NewBoolQuery()
		termQueries        []elastic.Query
		mustNotTermQueries []elastic.Query
		methodKeyword      = "method.keyword"
		httpModelKeyword   = "http_model.keyword"
		authKeyword        = "http_authorization.keyword"
	)

	// 验证时间范围
	if from == 0 || to == 0 || to <= from {
		return nil, fmt.Errorf("invalid time range: from=%d to=%d", from, to)
	}

	termQueries = append(termQueries,
		elastic.NewTermsQuery(methodKeyword, "POST"))

	query = query.Must(termQueries...)
	query = query.Must(elastic.NewRangeQuery("@timestamp").From(from).To(to).Format("epoch_millis"))

	mustNotTermQueries = append(mustNotTermQueries,
		elastic.NewTermsQuery(httpModelKeyword, "-"))
	mustNotTermQueries = append(mustNotTermQueries,
		elastic.NewTermsQuery(httpModelKeyword, ""))
	query = query.MustNot(mustNotTermQueries...)

	modelAgg := elastic.NewTermsAggregation().
		Field(httpModelKeyword).
		Size(500).
		OrderByKeyDesc().
		MinDocCount(1).
		Missing("N/A")

	authAgg := elastic.NewTermsAggregation().
		Field(authKeyword).
		Size(500).
		OrderByKeyDesc().
		MinDocCount(1)

	modelAgg = modelAgg.SubAggregation("authorizations", authAgg)

	aggsName := "nested_models_authorizations" // 更新聚合名称以反映新结构
	searchService := e.ESClient.Client.Search().
		Index(e.Index).
		Query(query).
		Size(0).
		IgnoreUnavailable(true).
		SearchType("query_then_fetch").
		Aggregation(aggsName, modelAgg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("ES查询失败: %v", err)
	}

	requestCountMap := make(map[string]map[string]int64)

	modelTerms, found := searchResult.Aggregations.Terms(aggsName)
	if !found {
		return requestCountMap, fmt.Errorf("未找到聚合结果")
	}

	for _, modelBucket := range modelTerms.Buckets {
		modelKey, ok := modelBucket.Key.(string)
		if !ok {
			modelKey = fmt.Sprintf("%v", modelBucket.Key)
		}

		authTerms, ok := modelBucket.Aggregations.Terms("authorizations")
		if !ok {
			log.Printf("模型分组 %s 中没有找到授权聚合", modelKey)
			continue
		}

		authMap := make(map[string]int64)
		for _, authBucket := range authTerms.Buckets {
			authKey, ok := authBucket.Key.(string)
			if !ok {
				authKey = fmt.Sprintf("%v", authBucket.Key)
			}
			authMap[authKey] = authBucket.DocCount
		}

		requestCountMap[modelKey] = authMap
	}

	return requestCountMap, nil
}
