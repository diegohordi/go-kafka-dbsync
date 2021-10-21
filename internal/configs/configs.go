package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type DBConfigurer interface {
	DSN() string
}

type KafkaConfigurer interface {
	DSN() string
	Topic() string
	Partition() int
}

type AppConfigurer interface {
	Port() int
}

type Configurer interface {
	DB() DBConfigurer
	Kafka() KafkaConfigurer
	App() AppConfigurer
}

type config struct {
	kafkaConfig
	dbConfig
	appConfig
}

type dbConfig struct {
	dsn string
}

func (d dbConfig) DSN() string {
	return d.dsn
}

type kafkaConfig struct {
	dsn       string
	topic     string
	partition int
}

func (c kafkaConfig) DSN() string {
	return c.dsn
}

func (c kafkaConfig) Topic() string {
	return c.topic
}

func (c kafkaConfig) Partition() int {
	return c.partition
}

type appConfig struct {
	port int
}

func (a appConfig) Port() int {
	return a.port
}

func (c config) DB() DBConfigurer {
	return c.dbConfig
}

func (c config) Kafka() KafkaConfigurer {
	return c.kafkaConfig
}

func (c config) App() AppConfigurer {
	return c.appConfig
}

// Load loads the given configuration file.
func Load(configPath string) (Configurer, error) {
	data := &config{}
	dbConf, err := loadDBConfig(configPath)
	if err != nil {
		return nil, err
	}
	data.dbConfig = *dbConf
	kafkaConf, err := loadKafkaConfig(configPath)
	if err != nil {
		return nil, err
	}
	data.kafkaConfig = *kafkaConf
	appConf, err := loadAppConfig(configPath)
	if err != nil {
		return nil, err
	}
	data.appConfig = *appConf
	return *data, nil
}

func loadAppConfig(configPath string) (*appConfig, error) {
	appConf := &appConfig{}
	appConf.port = 8080
	if port, err := strconv.Atoi(os.Getenv("APP_PORT")); err == nil {
		appConf.port = port
	}
	if configPath != "" {
		confDef := &struct {
			App struct {
				Port int    `json:"port"`
			} `json:"app"`
		}{}
		configFile, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("an occurred while loading config file: %w", err)
		}
		err = json.NewDecoder(configFile).Decode(confDef)
		if err != nil {
			return nil, fmt.Errorf("an occurred while parsing config file: %w", err)
		}
		appConf.port = confDef.App.Port
	}
	return appConf, nil
}

func loadDBConfig(configPath string) (*dbConfig, error) {
	dbConf := &dbConfig{}
	dbConf.dsn = os.Getenv("DATABASE_DSN")
	if configPath != "" {
		confDef := &struct {
			DB struct {
				DSN string `json:"dsn"`
			} `json:"db"`
		}{}
		configFile, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("an occurred while loading config file: %w", err)
		}
		err = json.NewDecoder(configFile).Decode(confDef)
		if err != nil {
			return nil, fmt.Errorf("an occurred while parsing config file: %w", err)
		}
		dbConf.dsn = confDef.DB.DSN
	}
	return dbConf, nil
}

func loadKafkaConfig(configPath string) (*kafkaConfig, error) {
	kafkaConf := &kafkaConfig{}
	kafkaConf.dsn = os.Getenv("KAFKA_DSN")
	kafkaConf.topic = os.Getenv("KAFKA_TOPIC")
	if partition, err := strconv.Atoi(os.Getenv("KAFKA_PARTITION")); err == nil {
		kafkaConf.partition = partition
	}
	if configPath != "" {
		confDef := &struct {
			Kafka struct {
				DSN       string `json:"dsn"`
				Topic     string `json:"topic"`
				Partition int    `json:"partition"`
			} `json:"kafka"`
		}{}
		configFile, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("an occurred while loading config file: %w", err)
		}
		err = json.NewDecoder(configFile).Decode(confDef)
		if err != nil {
			return nil, fmt.Errorf("an occurred while parsing config file: %w", err)
		}
		kafkaConf.dsn = confDef.Kafka.DSN
		kafkaConf.topic = confDef.Kafka.Topic
		kafkaConf.partition = confDef.Kafka.Partition
	}
	return kafkaConf, nil
}

func MustLoad(configPath string) Configurer {
	conf, err := Load(configPath)
	if err != nil {
		panic(err)
	}
	return conf
}
