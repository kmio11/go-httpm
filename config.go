package httpm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Default *Action `toml:"default,omitempty"`
	Rules   Rules   `toml:"rules"`
}

type Rules []RuleConfig

type RuleConfig struct {
	Condition Condition `toml:"condition"`
	Action    *Action   `toml:"action,omitempty"`
}

type Condition struct {
	Methods []string `toml:"method"`
	URL     string   `toml:"url"`
}

type Action struct {
	Type         ActionType `toml:"type"`
	ResponseFile string     `toml:"response_file,omitempty"`
	Response     string     `toml:"response,omitempty"`
}

type ActionType string

const (
	// ActionTypeMock represents an action type where the request is mocked.
	ActionTypeMock ActionType = "mock"
	// ActionTypePass represents an action type where the request is passed through without any modification.
	ActionTypePass ActionType = "pass"
	// ActionTypePanic represents an action type where the request causes a panic.
	ActionTypePanic ActionType = "panic"
)

func NewActionPass() *Action {
	return &Action{
		Type: ActionTypePass,
	}
}

func NewActionPanic() *Action {
	return &Action{
		Type: ActionTypePanic,
	}
}

func NewActionMock(response string) *Action {
	return &Action{
		Type:     ActionTypeMock,
		Response: response,
	}
}

func LoadConfigFile(file string) (*Config, error) {
	var config Config
	configFileDir := filepath.Dir(file)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := toml.DecodeFile(file, &config); err != nil {
		return nil, err
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	if err := loadResponseFiles(&config, configFileDir); err != nil {
		return nil, err
	}

	return &config, nil
}

func loadResponseFiles(config *Config, configFileDir string) error {
	for i, rule := range config.Rules {
		if rule.Action.ResponseFile != "" {
			responsePath := rule.Action.ResponseFile
			if !filepath.IsAbs(responsePath) {
				responsePath = filepath.Join(configFileDir, responsePath)
			}
			response, err := os.ReadFile(responsePath)
			if err != nil {
				return fmt.Errorf("failed to read response file %s: %w", responsePath, err)
			}
			config.Rules[i].Action.Response = string(response)
		}
	}
	return nil
}

func validateConfig(config *Config) error {
	for i, rule := range config.Rules {
		if err := validateRule(&rule); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}
	return nil
}

func validateRule(rule *RuleConfig) error {
	if err := validateCondition(&rule.Condition); err != nil {
		return fmt.Errorf("condition: %w", err)
	}
	if err := validateAction(rule.Action); err != nil {
		return fmt.Errorf("action: %w", err)
	}
	return nil
}

func validateCondition(_ *Condition) error {
	return nil
}

func validateAction(action *Action) error {
	if action == nil {
		return fmt.Errorf("action must be provided")
	}

	switch action.Type {
	case ActionTypeMock:
		if action.Response == "" && action.ResponseFile == "" {
			return fmt.Errorf("response or response_file must be provided for action type %s", action.Type)
		}
	case ActionTypePass:
		return nil
	case ActionTypePanic:
		return nil
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
	return nil
}
