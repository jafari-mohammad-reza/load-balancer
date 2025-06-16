package conf

import (
	"github.com/spf13/viper"
)

type Logger string

const (
	JSON  Logger = "json"
	KAFKA Logger = "kafka"
)

type Conf struct {
	Port    int         `mapstructure:"port"`
	Proxies []ProxyConf `mapstructure:"proxies"`
	Log     LogConf     `mapstructure:"log"`
	Kafka   KafkaConf   `mapstructure:"kafka"`
}

type ProxyConf struct {
	Port           int            `mapstructure:"port"`
	Host           string         `mapstructure:"host"`
	TLS            bool           `mapstructure:"tls"`
	Certificate    string         `mapstructure:"certificate"`
	CertificateKey string         `mapstructure:"certificate_key"`
	ClientCA       string         `mapstructure:"certificate_ca"`
	Locations      []LocationConf `mapstructure:"locations"`
}

type LocationConf struct {
	Path           string          `mapstructure:"path"`
	Algorithm      string          `mapstructure:"algorithm"`
	BackendServers []BackendServer `mapstructure:"backend_servers"`
}

type BackendServer struct {
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	Weight int    `mapstructure:"weight"`
}

type LogConf struct {
	Logger  Logger `mapstructure:"logger"`
	LogPath string `mapstructure:"log_path"`
}
type KafkaConf struct {
	Servers  string `mapstructure:"servers"`
	ClientId string `mapstructure:"client_id"`
	LogTopic string `mapstructure:"log_topic"`
}

func ReadConf() (*Conf, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.WatchConfig()
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	conf := &Conf{}
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}
	return conf, nil
}
