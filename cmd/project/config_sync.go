package project

import (
	"encoding/json"
	"fmt"

	adminSdk "github.com/haokeyingxiao/go-haoke-admin-api-sdk"

	"github.com/haokeyingxiao/haoke-cli/shop"
)

func readSystemConfig(ctx adminSdk.ApiContext, client *adminSdk.Client, salesChannelId *string) (*adminSdk.SystemConfigCollection, error) {
	c := adminSdk.Criteria{}
	c.Includes = map[string][]string{"system_config": {"id", "configurationKey", "configurationValue"}}

	c.Filter = []adminSdk.CriteriaFilter{
		{Type: adminSdk.SearchFilterTypeEquals, Field: "salesChannelId", Value: salesChannelId},
	}

	results, resp, err := client.Repository.SystemConfig.SearchAll(ctx, c)
	if err != nil {
		return nil, err
	}

	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	return results, nil
}

type ConfigSyncApplyer interface {
	Push(ctx adminSdk.ApiContext, client *adminSdk.Client, config *shop.Config, operation *ConfigSyncOperation) error
	Pull(ctx adminSdk.ApiContext, client *adminSdk.Client, config *shop.Config) error
}

func NewSyncApplyers() []ConfigSyncApplyer {
	return []ConfigSyncApplyer{SystemConfigSync{}, ThemeSync{}, MailTemplateSync{}, EntitySync{}}
}

type ConfigSyncOperation struct {
	Operations     Operation
	SystemSettings SystemConfig
	ThemeSettings  ThemeSettings
}

type ThemeSyncOperation struct {
	Id       string
	Name     string
	Settings map[string]adminSdk.ThemeConfigValue
}

type (
	Operation     map[string]adminSdk.SyncOperation
	SystemConfig  map[*string]map[string]interface{}
	ThemeSettings []ThemeSyncOperation
)

func (o ConfigSyncOperation) HasChanges() bool {
	return o.Operations.HasChanges() || o.SystemSettings.HasChanges() || o.ThemeSettings.HasChanges()
}

func (o Operation) HasChanges() bool {
	return len(o) > 0
}

func (t ThemeSettings) HasChanges() bool {
	for _, m := range t {
		if len(m.Settings) > 0 {
			return true
		}
	}

	return false
}

func (s SystemConfig) ToJson() string {
	text := ""

	for key, v := range s {
		if len(v) == 0 {
			continue
		}

		content, _ := json.Marshal(v)

		var k string

		if key == nil {
			k = `"null"`
		} else {
			k = fmt.Sprintf(`%q`, *key)
		}

		text += fmt.Sprintf(`%s: %s,`, k, content)
	}

	if text == "" {
		return "{}"
	}

	return fmt.Sprintf("{%s}", text[:len(text)-1])
}

func (s SystemConfig) HasChanges() bool {
	for _, m := range s {
		if len(m) > 0 {
			return true
		}
	}

	return false
}
