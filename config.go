package main 

type Config struct {
	ListenAddress string
	ListenPort 	int
}

func loadConfig() *Config {
	defaultConfig := &Config{
		ListenAddress: "0.0.0.0", 
		ListenPort: 23, 
	}

	return defaultConfig
}
