package account_api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gorilla/schema"
)

type ProducerEndpoint struct {
	c          *Client
	producerId string
}

func (e ProducerEndpoint) GetId() string {
	return e.producerId
}
func (c *Client) Producer(ctx context.Context) (*ProducerEndpoint, error) {
	r, err := c.NewAuthenticatedRequest(ctx, "GET", fmt.Sprintf("%s/producers/%s", ApiUrl, c.GetUserID()), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(r)
	if err != nil {
		return nil, err
	}

	var allocation companyAllocation
	if err := json.Unmarshal(body, &allocation); err != nil {
		return nil, fmt.Errorf("producer.profile: %v", err)
	}

	if !allocation.IsProducer {
		return nil, fmt.Errorf("this user is not unlocked as producer")
	}

	return &ProducerEndpoint{producerId: allocation.ProducerID, c: c}, nil
}

type companyAllocation struct {
	IsProducer bool   `json:"isProducer"`
	ProducerID string `json:"producerId"`
}

func (e ProducerEndpoint) Profile(ctx context.Context) (*Producer, error) {
	r, err := e.c.NewAuthenticatedRequest(ctx, "GET", fmt.Sprintf("%s/producers?producersId=%s", ApiUrl, e.GetId()), nil)
	if err != nil {
		return nil, err
	}

	body, err := e.c.doRequest(r)
	if err != nil {
		return nil, err
	}

	var producers []Producer
	if err := json.Unmarshal(body, &producers); err != nil {
		return nil, fmt.Errorf("my_profile: %v", err)
	}

	for _, profile := range producers {
		return &profile, nil
	}

	return nil, fmt.Errorf("cannot find a profile")
}

type Producer struct {
	Id      string `json:"id"`
	Prefix  string `json:"prefix"`
	Name    string `json:"name"`
	HaokeID string `json:"haokeId"`
}

type ListExtensionCriteria struct {
	Limit         int    `schema:"limit,omitempty"`
	Offset        int    `schema:"offset,omitempty"`
	OrderBy       string `schema:"orderBy,omitempty"`
	OrderSequence string `schema:"orderSequence,omitempty"`
	Search        string `schema:"search,omitempty"`
}

func (e ProducerEndpoint) Extensions(ctx context.Context, criteria *ListExtensionCriteria) ([]Extension, error) {
	encoder := schema.NewEncoder()
	form := url.Values{}
	form.Set("producerId", e.GetId())
	err := encoder.Encode(criteria, form)
	if err != nil {
		return nil, fmt.Errorf("list_extensions: %v", err)
	}

	r, err := e.c.NewAuthenticatedRequest(ctx, "GET", fmt.Sprintf("%s/plugins?%s", ApiUrl, form.Encode()), nil)
	if err != nil {
		return nil, err
	}

	body, err := e.c.doRequest(r)
	if err != nil {
		return nil, err
	}

	var extensions []Extension
	if err := json.Unmarshal(body, &extensions); err != nil {
		return nil, fmt.Errorf("list_extensions: %v", err)
	}

	return extensions, nil
}

func (e ProducerEndpoint) GetExtensionByName(ctx context.Context, name string) (*Extension, error) {
	criteria := ListExtensionCriteria{
		Search: name,
	}

	extensions, err := e.Extensions(ctx, &criteria)
	if err != nil {
		return nil, err
	}

	for _, ext := range extensions {
		if strings.EqualFold(ext.Name, name) {
			return e.GetExtensionById(ctx, ext.Id)
		}
	}

	return nil, fmt.Errorf("cannot find Extension by name %s", name)
}

func (e ProducerEndpoint) GetExtensionById(ctx context.Context, id int) (*Extension, error) {
	errorFormat := "GetExtensionById: %v"

	// Create it
	r, err := e.c.NewAuthenticatedRequest(ctx, "GET", fmt.Sprintf("%s/plugins/%d", ApiUrl, id), nil)
	if err != nil {
		return nil, fmt.Errorf(errorFormat, err)
	}

	body, err := e.c.doRequest(r)
	if err != nil {
		return nil, fmt.Errorf(errorFormat, err)
	}

	var extension Extension
	if err := json.Unmarshal(body, &extension); err != nil {
		return nil, fmt.Errorf(errorFormat, err)
	}

	return &extension, nil
}

type Extension struct {
	Id             int    `json:"id"`
	ProducerId     string `json:"producerId"`
	Type           string `json:"type"`
	Name           string `json:"name"`
	StandardLocale Locale `json:"standardLocale"`
	Infos          []*struct {
		Id                 int          `json:"id"`
		Locale             Locale       `json:"locale"`
		Name               string       `json:"name"`
		Description        string       `json:"description"`
		InstallationManual string       `json:"installationManual"`
		ShortDescription   string       `json:"shortDescription"`
		Highlights         string       `json:"highlights"`
		Features           string       `json:"features"`
		Tags               []StoreTag   `json:"tags"`
		Videos             []StoreVideo `json:"videos"`
		Faqs               []StoreFaq   `json:"faqs"`
	} `json:"infos"`
	PriceModels                         *StorePrice       `json:"priceModels"`
	Variants                            []interface{}     `json:"variants"`
	Categories                          []StoreCategory   `json:"categories"`
	Category                            *StoreCategory    `json:"selectedFutureCategory"`
	Addons                              []interface{}     `json:"addons"`
	AutomaticBugfixVersionCompatibility bool              `json:"automaticBugfixVersionCompatibility"`
	ProductType                         *StoreProductType `json:"productType"`
	Status                              struct {
		Name string `json:"name"`
	} `json:"status"`
	IconURL                               string `json:"iconUrl"`
	IsCompatibleWithLatestShopwareVersion bool   `json:"isCompatibleWithLatestShopwareVersion"`
}

type StorePrice struct {
	Type  string  `json:"type"`
	Money float32 `json:"money"`
}

type CreateExtensionRequest struct {
	Name       string `json:"name,omitempty"`
	Generation struct {
		Name string `json:"name"`
	} `json:"generation"`
	ProducerID string `json:"producerId"`
}

const (
	GenerationThemes   = "themes"
	GenerationApps     = "apps"
	GenerationPlatform = "platform"
)

func (e ProducerEndpoint) CreateExtension(ctx context.Context, newExtension CreateExtensionRequest) (*Extension, error) {
	requestBody, err := json.Marshal(newExtension)
	if err != nil {
		return nil, err
	}

	// Create it
	r, err := e.c.NewAuthenticatedRequest(ctx, "POST", fmt.Sprintf("%s/plugins", ApiUrl), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	body, err := e.c.doRequest(r)
	if err != nil {
		return nil, err
	}

	var extension Extension
	if err := json.Unmarshal(body, &extension); err != nil {
		return nil, fmt.Errorf("create_extension: %v", err)
	}
	return &extension, nil
}

func (e ProducerEndpoint) UpdateExtension(ctx context.Context, extension *Extension) error {
	requestBody, err := json.Marshal(extension)
	if err != nil {
		return err
	}

	// Patch the name
	r, err := e.c.NewAuthenticatedRequest(ctx, "PUT", fmt.Sprintf("%s/plugins/%d", ApiUrl, extension.Id), bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	_, err = e.c.doRequest(r)

	return err
}

func (e ProducerEndpoint) DeleteExtension(ctx context.Context, id string) error {
	r, err := e.c.NewAuthenticatedRequest(ctx, "DELETE", fmt.Sprintf("%s/plugins/%s", ApiUrl, id), nil)
	if err != nil {
		return err
	}

	_, err = e.c.doRequest(r)

	return err
}

func (e ProducerEndpoint) GetSoftwareVersions(ctx context.Context, generation string) (*SoftwareVersionList, error) {
	errorFormat := "shopware_versions: %v"
	r, err := e.c.NewAuthenticatedRequest(ctx, "GET", fmt.Sprintf("%s/pluginstatics/softwareVersions?filter=[{\"property\":\"pluginGeneration\",\"value\":\"%s\"},{\"property\":\"includeNonPublic\",\"value\":\"1\"}]", ApiUrl, generation), nil)
	if err != nil {
		return nil, fmt.Errorf(errorFormat, err)
	}

	body, err := e.c.doRequest(r)
	if err != nil {
		return nil, fmt.Errorf(errorFormat, err)
	}

	var versions SoftwareVersionList

	err = json.Unmarshal(body, &versions)

	if err != nil {
		return nil, fmt.Errorf(errorFormat, err)
	}

	return &versions, nil
}

type SoftwareVersion struct {
	Name       string `json:"version"`
	Selectable bool   `json:"selectable"`
}

type Locale struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type StoreCategory struct {
	Id          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parent      interface{} `json:"parent"`
	Position    int         `json:"position"`
	Public      bool        `json:"public"`
	Visible     bool        `json:"visible"`
	Suggested   bool        `json:"suggested"`
	Applicable  bool        `json:"applicable"`
	Details     interface{} `json:"details"`
	Active      bool        `json:"active"`
}

type StoreTag struct {
	Name string `json:"name"`
}

type StoreVideo struct {
	URL string `json:"url"`
}

type StoreProductType struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StoreFaq struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type ExtensionGeneralInformation struct {
	Categories       []StoreCategory `json:"categories"`
	FutureCategories []StoreCategory `json:"futureCategories"`
	Addons           interface{}     `json:"addons"`
	Generations      []struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"generations"`
	ActivationStatus []struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"activationStatus"`
	ApprovalStatus []struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"approvalStatus"`
	LifecycleStatus []struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"lifecycleStatus"`
	BinaryStatus []struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"binaryStatus"`
	Locales  []Locale `json:"locales"`
	Licenses []struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"licenses"`
	PriceModels          []interface{}       `json:"priceModels"`
	SoftwareVersions     SoftwareVersionList `json:"softwareVersions"`
	DemoTypes            interface{}         `json:"demoTypes"`
	ProductTypes         []StoreProductType  `json:"productTypes"`
	ReleaseRequestStatus interface{}         `json:"releaseRequestStatus"`
}

func (e ProducerEndpoint) GetExtensionGeneralInfo(ctx context.Context) (*ExtensionGeneralInformation, error) {
	r, err := e.c.NewAuthenticatedRequest(ctx, "GET", fmt.Sprintf("%s/pluginstatics/all", ApiUrl), nil)
	if err != nil {
		return nil, fmt.Errorf("GetExtensionGeneralInfo: %v", err)
	}

	body, err := e.c.doRequest(r)
	if err != nil {
		return nil, fmt.Errorf("GetExtensionGeneralInfo: %v", err)
	}

	var info *ExtensionGeneralInformation

	err = json.Unmarshal(body, &info)

	if err != nil {
		return nil, fmt.Errorf("shopware_versions: %v", err)
	}

	return info, nil
}
