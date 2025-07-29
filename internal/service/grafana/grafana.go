package grafana

import (
	"encoding/json"
	"errors"
	"monitor/internal/models"
	"net/url"

	grafana "github.com/grafana/grafana-api-golang-client"
)

type GrafanaClient struct {
	client *grafana.Client
}

func NewGrafanaClient(grafanaUrl, username, password, apikey string) *GrafanaClient {

	// 创建客户端
	client := &grafana.Client{}
	var err error
	if username != "" && password != "" {
		userInfo := url.UserPassword(username, password)
		client, err = grafana.New(grafanaUrl, grafana.Config{
			BasicAuth: userInfo,
		})
	}
	if apikey != "" {
		// fmt.Println("use grafana api key")
		client, err = grafana.New(grafanaUrl, grafana.Config{
			APIKey: apikey,
		})
	}
	if err != nil {
		panic(err)
	}
	return &GrafanaClient{client: client}
}

func (c *GrafanaClient) GetPanelMapping(uid string, title string) (map[string]models.Item, error) {
	dashboard, err := c.client.DashboardByUID(uid)
	if err != nil {
		panic(err)
	}
	panelsBytes, err := json.Marshal(dashboard.Model["panels"])
	if err != nil {
		return nil, err
	}
	panels := make([]models.Panel, 0)
	err = json.Unmarshal(panelsBytes, &panels)
	if err != nil {
		return nil, err
	}
	for _, panel := range panels {
		if panel.Title != title {
			continue
		}
		if len(panel.FieldConfig.Defaults.Mappings) == 0 {
			err := errors.New("no mappings")
			panic(err)
		}
		return panel.FieldConfig.Defaults.Mappings[0].Options, nil

	}
	return nil, nil
}
