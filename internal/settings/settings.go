package settings

import (
	"fmt"
	"gomificator/internal/constnats"
	"gomificator/internal/utils"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var weekdayMap = map[string]time.Weekday{
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
	"sun": time.Sunday,
}

type Config struct {
	// PomoConfig PomodoroConfig     `yaml:"pomodoro"`
	DayTypes           map[string]DayType       `yaml:"daytypes"`
	CalendarRaw        map[string]string        `yaml:"calendar"`
	Celendar           map[time.Weekday]DayType `yaml:"-"`
	AlwaysRestAfterStr string                   `yaml:"alwaysrestafter"`
	AlwaysRestAfter    time.Time                `yaml:"-"`
	AutoImport         AutoImportConfig         `yaml:"autoimport"`
}

func (c *Config) Validate() error {

	// if err := c.PomoConfig.Validate(); err != nil {
	// 	return fmt.Errorf("pomodoro: %w", err)
	// }

	for dayTypeName, dayType := range c.DayTypes {
		if err := dayType.Validate(); err != nil {
			return fmt.Errorf("day type %q: %w", dayTypeName, err)
		}
		dayType.Name = dayTypeName
		c.DayTypes[dayTypeName] = dayType
	}

	c.Celendar = make(map[time.Weekday]DayType)
	for weekdayStr, dayTypeStr := range c.CalendarRaw {
		weekday, ok := weekdayMap[weekdayStr]
		if !ok {
			return fmt.Errorf("unknown weekday string: %s", weekdayStr)
		}
		dayType, ok := c.DayTypes[dayTypeStr]
		if !ok {
			return fmt.Errorf("unknown day type string: %s", dayTypeStr)
		}
		c.Celendar[weekday] = dayType
	}

	alwaysRestAfter, err := time.Parse(constnats.TimeLayout, c.AlwaysRestAfterStr)
	if err != nil {
		return fmt.Errorf("time parse: %w", err)
	}
	c.AlwaysRestAfter = alwaysRestAfter

	if err := c.AutoImport.Validate(); err != nil {
		return fmt.Errorf("autoimport: %w", err)
	}
	return nil
}

// type PomodoroConfig struct {
// 	PomoLength       int `yaml:"pomolength" validate:"gte=1,lte=60"`
// 	RestLength       int `yaml:"pomorest" validate:"gte=1,lte=60"`
// 	LongRestLength   int `yaml:"longrestlength" validate:"gte=1,lte=60"`
// 	PomosTilLongRest int `yaml:"pomostillongrest" validate:"gte=1,lte=60"`
// }
//
// func (c *PomodoroConfig) Validate() error {
// 	validate := validator.New(validator.WithRequiredStructEnabled())
// 	if err := validate.Struct(c); err != nil {
// 		return fmt.Errorf("validate struct: %w", err)
// 	}
// 	return nil
// }

type DayType struct {
	Name       string         `yaml:"-"`
	FocusGoals []FocusDayGoal `yaml:"focusgoals"`
}

func (d *DayType) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(d); err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}

	var errs []error
	for idx := range d.FocusGoals {
		if err := d.FocusGoals[idx].Validate(); err != nil {
			errs = append(errs, fmt.Errorf("focus goals: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("day type validation errors: %v", errs)
	}

	return nil
}

type FocusDayGoal struct {
	RestAfterStr string    `yaml:"restafter" validate:"required"`
	RestAfter    time.Time `yaml:"-"`

	Minutes int `yaml:"minutes" validate:"gte=0,lte=1440"`
	Count   int `yaml:"count" validate:"gte=0,lte=1440"`

	MedalStr string          `yaml:"medal" validate:"required"`
	Medal    constnats.Medal `yaml:"-"`
}

func (f *FocusDayGoal) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(f); err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}

	medal, err := constnats.LoadMedal(f.MedalStr)
	if err != nil {
		return fmt.Errorf("load medal: %w", err)
	}
	f.Medal = medal

	time, err := time.Parse(constnats.TimeLayout, f.RestAfterStr)
	if err != nil {
		return fmt.Errorf("parse rest after time: %w", err)
	}
	f.RestAfter = time

	return nil
}

func newDefaultConfig() *Config {
	return &Config{
		// PomoConfig: PomodoroConfig{
		// 	PomoLength:       25,
		// 	RestLength:       5,
		// 	LongRestLength:   15,
		// 	PomosTilLongRest: 4,
		// },
	}
}

type AutoImportConfig struct {
	EveryStr string        `yaml:"every" validate:"required"`
	Every    time.Duration `yaml:"-"`
	Path     string        `yaml:"path" validate:"required"`
}

func (a *AutoImportConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(a); err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}

	d, err := time.ParseDuration(a.EveryStr)
	if err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}
	if d <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	a.Every = d
	return nil
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
