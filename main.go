package main

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"os"
	"path"

	"github.com/josherick/smart-home-control/logger"
	"github.com/josherick/smart-home-control/plug_controller"
	"github.com/josherick/smart-home-control/server"
)

type Config struct {
	KasaDirectory string `yaml:"kasa_dir" envconfig:"KASA_DIR"`
	SystemName    string `yaml:"system_name" envconfig:"SYSTEM_NAME"`
	Server        struct {
		Port int `yaml:"port" envconfig:"SERVER_PORT"`
	} `yaml:"server"`
	Logging struct {
		Directory string `yaml:"directory" envconfig:"LOG_DIRECTORY"`
	} `yaml:"logging"`
	Email struct {
		From     string `yaml:"from" envconfig:"EMAIL_FROM"`
		Password string `yaml:"password" envconfig:"EMAIL_PASSWORD"`
		To       string `yaml:"to" envconfig:"EMAIL_TO"`
		SmtpHost string `yaml:"smtp_host" envconfig:"SMTP_HOST"`
		SmtpPort string `yaml:"smtp_port" envconfig:"SMTP_PORT"`
	} `yaml:"email"`
	Devices struct {
		Plugs []struct {
			ID     string `yaml:"id"`
			IPAddr string `yaml:"ip_addr"`
		} `yaml:"plugs"`
		Sensors []struct {
			ID                  string `yaml:"id"`
			CorrespondingPlugIP string `yaml:"corresponding_plug_ip"`
		} `yaml:"sensors"`
	} `yaml:"devices"`
}

func main() {
	var config Config
	readFile(&config)
	readEnv(&config)

	plugCtrl := plug_controller.New(config.KasaDirectory, config.Devices.Sensors)
	mailClient := logger.NewEmailClient(
		config.Email.From,
		config.Email.Password,
		[]string{config.Email.To},
		config.Email.SmtpHost,
		config.Email.SmtpPort,
	)
	infoWriter := logger.NewFileWriter(path.Join(config.Logging.Directory, "info"))
	accessWriter := logger.NewFileWriter(path.Join(config.Logging.Directory, "access"))
	logger := logger.NewLogger(
		config.SystemName,
		infoWriter,
		accessWriter,
		mailClient,
	)

	server := server.New(config.Server.Port, logger, plugCtrl, infoWriter)
	server.Serve()
}

func readFile(config *Config) error {
	f, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)
	return err
}

func readEnv(config *Config) error {
	return envconfig.Process("", config)
}
