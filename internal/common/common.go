package common

import (
	"fmt"
	"monitor/config"
	"net/url"
)

type Lang string

const (
	LangZh Lang = "zh"
	LangEn Lang = "en"
)

func BuildGrafanaDashboardURL(uid, path string, vars map[string][]string, queries map[string][]string, lang Lang) string {
	u := url.URL{
		Path: fmt.Sprintf("/ui/insight-grafana/d/%s/%s", uid, path),
	}

	q := u.Query()
	for k, v := range vars {
		k = "var-" + k
		for _, item := range v {
			q.Add(k, item)
		}
	}

	for k, v := range queries {
		for _, item := range v {
			q.Add(k, item)
		}
	}

	if len(q.Get("kiosk")) == 0 {
		q.Add("kiosk", "tv")
	}

	qq, _ := url.ParseQuery(q.Encode()) // copy
	qq.Add("var-lang", string(lang))
	u.RawQuery = qq.Encode()
	return u.String()
}

func GenerateDocURL(cfg config.KibanaConfig, docID string) (string, error) {
	if cfg.IndexPatternID == "" || docID == "" {
		return "", fmt.Errorf("缺少必要参数")
	}
	index := config.GetEsConfig().Index
	indexId := config.GetKibanaConfig().IndexPatternID

	return fmt.Sprintf("/app/discover#/doc/%s/%s?id=%s", indexId, index, docID), nil
}
