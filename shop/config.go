package shop

import (
	"fmt"
	"os"
	"strings"

	"dario.cat/mergo"

	"github.com/doutorfinancas/go-mad/core"
	"github.com/google/uuid"
	adminSdk "github.com/haokeyingxiao/go-haoke-admin-api-sdk"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AdditionalConfigs []string        `yaml:"include,omitempty"`
	URL               string          `yaml:"url"`
	Build             *ConfigBuild    `yaml:"build,omitempty"`
	AdminApi          *ConfigAdminApi `yaml:"admin_api,omitempty"`
	ConfigDump        *ConfigDump     `yaml:"dump,omitempty"`
	Sync              *ConfigSync     `yaml:"sync,omitempty"`
	foundConfig       bool
}

type ConfigBuild struct {
	DisableAssetCopy      bool     `yaml:"disable_asset_copy,omitempty"`
	RemoveExtensionAssets bool     `yaml:"remove_extension_assets,omitempty"`
	KeepExtensionSource   bool     `yaml:"keep_extension_source,omitempty"`
	KeepSourceMaps        bool     `yaml:"keep_source_maps,omitempty"`
	CleanupPaths          []string `yaml:"cleanup_paths,omitempty"`
	Browserslist          string   `yaml:"browserslist,omitempty"`
	ExcludeExtensions     []string `yaml:"exclude_extensions,omitempty"`
}

type ConfigAdminApi struct {
	ClientId        string `yaml:"client_id,omitempty"`
	ClientSecret    string `yaml:"client_secret,omitempty"`
	Username        string `yaml:"username,omitempty"`
	Password        string `yaml:"password,omitempty"`
	DisableSSLCheck bool   `yaml:"disable_ssl_check,omitempty"`
}

type ConfigDump struct {
	Rewrite map[string]core.Rewrite `yaml:"rewrite,omitempty"`
	NoData  []string                `yaml:"nodata,omitempty"`
	Ignore  []string                `yaml:"ignore,omitempty"`
	Where   map[string]string       `yaml:"where,omitempty"`
}

type ConfigSync struct {
	Config       []ConfigSyncConfig `yaml:"config"`
	Theme        []ThemeConfig      `yaml:"theme"`
	MailTemplate []MailTemplate     `yaml:"mail_template"`
	Entity       []EntitySync       `yaml:"entity"`
}

type ConfigSyncConfig struct {
	SalesChannel *string                `yaml:"sales_channel,omitempty"`
	Settings     map[string]interface{} `yaml:"settings"`
}

type ThemeConfig struct {
	Name     string                               `yaml:"name"`
	Settings map[string]adminSdk.ThemeConfigValue `yaml:"settings"`
}

type MailTemplate struct {
	Id           string                    `yaml:"id"`
	Translations []MailTemplateTranslation `yaml:"translations"`
}

type EntitySync struct {
	Entity  string                 `yaml:"entity"`
	Exists  *[]interface{}         `yaml:"exists"`
	Payload map[string]interface{} `yaml:"payload"`
}

type MailTemplateTranslation struct {
	Language     string      `yaml:"language"`
	SenderName   string      `yaml:"sender_name"`
	Subject      string      `yaml:"subject"`
	HTML         string      `yaml:"html"`
	Plain        string      `yaml:"plain"`
	CustomFields interface{} `yaml:"custom_fields"`
}

func ReadConfig(fileName string, allowFallback bool) (*Config, error) {
	config := &Config{foundConfig: false}

	_, err := os.Stat(fileName)

	if os.IsNotExist(err) {
		if allowFallback {
			return fillEmptyConfig(config), nil
		}

		return nil, fmt.Errorf("cannot find project configuration file \"%s\", use shopware-cli project config init to create one", fileName)
	}

	if err != nil {
		return nil, err
	}

	fileHandle, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("ReadConfig: %v", err)
	}

	config.foundConfig = true

	substitutedConfig := os.ExpandEnv(string(fileHandle))
	err = yaml.Unmarshal([]byte(substitutedConfig), &config)

	if len(config.AdditionalConfigs) > 0 {
		for _, additionalConfigFile := range config.AdditionalConfigs {
			additionalConfig, err := ReadConfig(additionalConfigFile, allowFallback)
			if err != nil {
				return nil, fmt.Errorf("error while reading included config: %s", err.Error())
			}

			err = mergo.Merge(additionalConfig, config, mergo.WithOverride, mergo.WithSliceDeepCopy)
			if err != nil {
				return nil, fmt.Errorf("error while merging included config: %s", err.Error())
			}

			config = additionalConfig
		}
	}

	if err != nil {
		return nil, fmt.Errorf("ReadConfig: %v", err)
	}

	return fillEmptyConfig(config), nil
}

func fillEmptyConfig(c *Config) *Config {
	if c.Build == nil {
		c.Build = &ConfigBuild{}
	}

	return c
}

func (c Config) IsFallback() bool {
	return !c.foundConfig
}

func NewUuid() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
