package settings

import (
	"fmt"
	"gomificator/internal/utils"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	PomoConfig PomodoroConfig `yaml:"pomodoro"`
}

func (c *Config) Validate() error {
	if err := c.PomoConfig.Validate(); err != nil {
		return fmt.Errorf("pomodoro: %w", err)
	}

	return nil
}

type PomodoroConfig struct {
	PomoLength       int `yaml:"pomolength" validate:"gte=1,lte=60"`
	RestLength       int `yaml:"pomorest" validate:"gte=1,lte=60"`
	LongRestLength   int `yaml:"longrestlength" validate:"gte=1,lte=60"`
	PomosTilLongRest int `yaml:"pomostillongrest" validate:"gte=1,lte=60"`
}

func (c *PomodoroConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}
	return nil
}
func newDefaultConfig() *Config {
	return &Config{
		PomoConfig: PomodoroConfig{
			PomoLength:       25,
			RestLength:       5,
			LongRestLength:   15,
			PomosTilLongRest: 4,
		},
	}
}

func initConfigFile(confPath string) (*Config, error) {
	cfg := newDefaultConfig()

	if err := SaveConfig(&confPath, cfg); err != nil {
		return nil, fmt.Errorf("save config %w", err)
	}

	return cfg, nil

}

func GetDefaultConfigPath() (string, error) {
	configDir, err := utils.GetAppDataLocation()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}
	configPath := filepath.Join(configDir, "settings.yaml")

	return configPath, nil
}

func LoadConfig(confPath *string) (*Config, error) {

	if confPath == nil {
		cp, err := GetDefaultConfigPath()
		if err != nil {
			return nil, fmt.Errorf("default coinfig path: %w", err)
		}
		confPath = &cp
	}

	if _, err := os.Stat(*confPath); os.IsNotExist(err) {
		cfg, err := initConfigFile(*confPath)

		if err != nil {
			return nil, fmt.Errorf("can't init config file: %w", err)
		}
		return cfg, nil
	}

	var cfg Config

	confFileData, err := os.ReadFile(*confPath)
	if err != nil {
		return nil, fmt.Errorf("can't read config file: %w", err)
	}

	if err = yaml.Unmarshal(confFileData, &cfg); err != nil {
		return nil, fmt.Errorf("can't unmarshal data from config file: %w", err)
	}

	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return &cfg, nil
}

func SaveConfig(confPath *string, cfg *Config) error {
	data, err := yaml.Marshal(*cfg)
	if err != nil {
		return fmt.Errorf("config saving: %w", err)
	}

	if err = writeToFile(*confPath, data); err != nil {
		return fmt.Errorf("config saving: %w", err)
	}

	return nil
}

func writeToFile(filename string, data []byte) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}
