package extension

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"io/ioutil"
	"os"
	"strings"
)

type PlatformPlugin struct {
	path                      string
	name                      string
	shopwareVersionConstraint version.Constraints
}

func newPlatformPlugin(path string) (*PlatformPlugin, error) {
	composerJsonFile := fmt.Sprintf("%s/composer.json", path)
	if _, err := os.Stat(composerJsonFile); err != nil {
		return nil, err
	}

	jsonFile, err := ioutil.ReadFile(composerJsonFile)

	if err != nil {
		return nil, err
	}

	var composerJson platformComposerJson
	err = json.Unmarshal(jsonFile, &composerJson)

	if err != nil {
		return nil, err
	}

	parts := strings.Split(composerJson.Extra.ShopwarePluginClass, "\\")
	shopwareConstraintString, ok := composerJson.Require["shopware/core"]

	if !ok {
		return nil, fmt.Errorf("require.shopware/core is required")
	}

	shopwareConstraint, err := version.NewConstraint(shopwareConstraintString)

	if err != nil {
		return nil, err
	}

	extension := PlatformPlugin{
		path:                      path,
		name:                      parts[len(parts)-1],
		shopwareVersionConstraint: shopwareConstraint,
	}

	return &extension, nil
}

type platformComposerJson struct {
	Name        string   `json:"name"`
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Type        string   `json:"type"`
	License     string   `json:"license"`
	Authors     []struct {
		Name     string `json:"name"`
		Homepage string `json:"homepage"`
	} `json:"authors"`
	Require map[string]string `json:"require"`
	Extra   struct {
		ShopwarePluginClass string            `json:"shopware-plugin-class"`
		Label               map[string]string `json:"label"`
		Description         map[string]string `json:"description"`
		ManufacturerLink    map[string]string `json:"manufacturerLink"`
		SupportLink         map[string]string `json:"supportLink"`
	} `json:"extra"`
	Autoload struct {
		Psr4 map[string]string `json:"psr-4"`
	} `json:"autoload"`
}

func (p PlatformPlugin) GetName() string {
	return p.name
}

func (p PlatformPlugin) GetShopwareVersionConstraint() version.Constraints {
	return p.shopwareVersionConstraint
}

func (p PlatformPlugin) GetType() string {
	return "platform"
}