package config

import (
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
	"log"
	"os"
	"time"
)

type ServerConfig struct {
	Port int `yaml:"port"`
}

type GrafanaQueryConfig struct {
	GrafanaDashBoard   string `mapstructure:"grafanadashboard"`
	ClusterBaseURL     string `mapstructure:"clusterBaseURL"`
	Mock               int    `mapstructure:"mock"`
	InsecureSkipVerify bool   `mapstructure:"insecureSkipVerify"`
	Token              string `mapstructure:"token"`
}

type KibanaConfig struct {
	IndexPatternID string `mapstructure:"index_pattern_id"`
	Url            string `mapstructure:"url"`
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

// MySQL DB 配置
type DBConfig struct {
	DBType               string        `yaml:"dbType"`               // 数据库类型，默认 mysql
	DSN                  string        `yaml:"dsn"`                  // data source name, e.g. root:123456@tcp(127.0.0.1:3306)/hydra
	MaxIdleConns         int           `yaml:"maxIdleConns"`         // 最大空闲连接数
	MaxOpenConns         int           `yaml:"maxOpenConns"`         // 最大连接数
	AutoMigrate          bool          `yaml:"autoMigrate"`          // 自动建表，补全缺失字段，初始化数据
	Debug                bool          `yaml:"debug"`                // 是否开启调试模式
	CacheFlag            bool          `yaml:"cacheFlag"`            // 是否开启查询缓存
	CacheExpiration      time.Duration `yaml:"cacheExpiration"`      // 缓存过期时间
	CacheCleanupInterval time.Duration `yaml:"cacheCleanupInterval"` // 缓存清理时间间隔
}

var (
	Gc         GrafanaQueryConfig
	ServerPort ServerConfig
	ModelFP    map[string]float64
	Es         ESConfig
	Grafana    GrafanaConfig
	Kibana     KibanaConfig
	DbConfig   DBConfig
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

	if err := newViper.UnmarshalKey("mysql", &DbConfig); err != nil {
		log.Println("Error reading config file, %s", err)
		return err
	}

	return nil
}
func GetDBConfig() *DBConfig {

	if DbConfig.MaxIdleConns <= 0 {
		DbConfig.MaxIdleConns = 10
	}
	if DbConfig.MaxOpenConns <= 0 {
		DbConfig.MaxOpenConns = 20
	}
	if DbConfig.CacheExpiration == 0 {
		DbConfig.CacheExpiration = 5 * time.Minute
	}
	if DbConfig.CacheCleanupInterval == 0 {
		DbConfig.CacheCleanupInterval = 10 * time.Minute
	}
	return &DbConfig
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
