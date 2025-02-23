package config

type ApiConfig struct {
	Webhook string `yaml:"webhook"`
}

func (a *ApiConfig) setDefault() {
	return
}

func (a *ApiConfig) check() (err ConfigError) {
	return nil
}
