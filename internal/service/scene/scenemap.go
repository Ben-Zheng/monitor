package scene

import (
	"monitor/config"
	"monitor/internal/service/grafana"
	"monitor/util"
	"strings"
)

type IGrafanaService interface {
	GenerateApiSixScenarioKeyMap() (map[string]string, error)
	GenerateApiSixScenarioKeyMapMock() (map[string]string, error)
}

type GrafanaService struct {
	GrafanaClient                   *grafana.GrafanaClient
	ModelRequestDashboardUid        string
	ModelRequestDashboardPanelTitle string
}

func NewGrafanaService(appConfig *config.GrafanaConfig) IGrafanaService {
	return &GrafanaService{
		GrafanaClient:                   grafana.NewGrafanaClient(appConfig.URL, appConfig.Username, appConfig.Password, appConfig.APIKey),
		ModelRequestDashboardUid:        appConfig.ModelRequestDashboardUid,
		ModelRequestDashboardPanelTitle: appConfig.ModelRequestDashboardPanelTitle,
	}
}

func (s *GrafanaService) GenerateApiSixScenarioKeyMap() (map[string]string, error) {
	apiSixScenarioKeyMap := make(map[string]string, 0)
	consumers, err := s.GrafanaClient.GetPanelMapping(s.ModelRequestDashboardUid, s.ModelRequestDashboardPanelTitle)
	if err != nil {
		return nil, nil
	}
	for key, consumer := range consumers {
		text := util.ProcessSceneString(consumer.Text)
		apiSixScenarioKeyMap[key] = text
		//fmt.Printf("%s--%s\n", consumer.Text, key)
	}
	//sceneLabel:scene
	return apiSixScenarioKeyMap, nil
}

func (s *GrafanaService) GenerateApiSixScenarioKeyMapMock() (map[string]string, error) {
	apiSixScenarioKeyMap := make(map[string]string, 0)
	consumers, err := s.GrafanaClient.GetPanelMapping(s.ModelRequestDashboardUid, s.ModelRequestDashboardPanelTitle)
	if err != nil {
		return nil, nil
	}
	for key, consumer := range consumers {
		//text := util.ProcessSceneString(consumer.Text)
		text := strings.ToLower(consumer.Text)
		apiSixScenarioKeyMap[text] = key
		//fmt.Printf("%s--%s\n", consumer.Text, key)
	}
	//log.Println("====================apiSixScenarioKeyMap", apiSixScenarioKeyMap)
	//sceneLabel:scene
	return apiSixScenarioKeyMap, nil
}
