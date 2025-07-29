package config

import (
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
	"log"
	"os"
)

type ServerConfig struct {
	Port int `yaml:"port"`
}

type GrafanaQueryConfig struct {
	RefID              string `mapstructure:"ref_id"`
	DatasourceID       int    `mapstructure:"datasource_id"`
	IntervalMs         int64  `mapstructure:"interval_ms"`
	MaxDataPoints      int64  `mapstructure:"max_data_points"`
	ClusterBaseURL     string `mapstructure:"clusterBaseURL"`
	TimeInterval       int64  `mapstructure:"timeinterval"`
	Mock               int    `mapstructure:"mock"`
	InsecureSkipVerify bool   `mapstructure:"insecureSkipVerify"`
	Token              string `mapstructure:"token"`
}

type KibanaConfig struct {
	IndexPatternID string `mapstructure:"index_pattern_id"`
}

type ESConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Index    string `yaml:"index"`
	Mock     bool   `yaml:"mock"`
}

type GrafanaConfig struct {
	URL                             string `yaml:"url"`
	Username                        string `yaml:"username"`
	Password                        string `yaml:"password"`
	APIKey                          string `yaml:"apikey"`
	ModelRequestDashboardUid        string `yaml:"modelRequestDashboardUid"`
	ModelRequestDashboardPanelTitle string `yaml:"modelRequestDashboardPanelTitle"`
}

var (
	Gc         GrafanaQueryConfig
	ServerPort ServerConfig
	ModelFP    map[string]float64
	Es         ESConfig
	Grafana    GrafanaConfig
	Kibana     KibanaConfig
)

func InitConfig() error {
	newViper := viper.NewWithOptions(viper.KeyDelimiter("::"))
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		path, err := os.Getwd()
		if err != nil {
			return xerrors.Errorf(": %w", err)
		}
		configPath = path + "/etc"
	}
	newViper.AddConfigPath(configPath)
	newViper.SetConfigName("config")
	newViper.SetConfigType("yaml")

	if err := newViper.ReadInConfig(); err != nil {
		log.Println("Error reading config file, %s", err)
		return err
	}

	if err := newViper.UnmarshalKey("grafana_query", &Gc); err != nil {
		log.Println("Error reading config file, %s", err)
		return err
	}
	if err := newViper.UnmarshalKey("es", &Es); err != nil {
		log.Println("Error reading config file, %s", err)
		return err
	}
	if err := newViper.UnmarshalKey("server", &ServerPort); err != nil {
		log.Println("Error reading config file, %s", err)
		return err
	}
	if err := newViper.UnmarshalKey("devices", &ModelFP); err != nil {

		log.Println("Error reading config file, %s", err)
		return err
	}

	if err := newViper.UnmarshalKey("grafana", &Grafana); err != nil {
		log.Println("Error reading config file, %s", err)
		return err
	}

	if err := newViper.UnmarshalKey("kibana", &Kibana); err != nil {
		log.Println("Error reading config file, %s", err)
		return err
	}
	return nil
}
func GetFPRule() map[string]float64 {
	return ModelFP
}

func GetGrafanaQueryConfig() GrafanaQueryConfig {
	return Gc
}
func GetEsConfig() ESConfig {
	return Es
}

func GetServerConfig() ServerConfig {
	return ServerPort
}
func GetKibanaConfig() KibanaConfig {
	return Kibana
}

func GetGrafanaConfig() GrafanaConfig {
	return Grafana
}
