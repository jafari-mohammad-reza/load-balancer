package conf

import "github.com/spf13/viper"

type Conf struct {
	Port           int    `mapstructure:"port"`
	Algorithm      string `mapstructure:"algorithm"`
	BackendServers []BackendServer `mapstructure:"backend_servers"`
}
type BackendServer struct {
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	Weight int    `mapstructure:"weight"`
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
