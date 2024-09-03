package traefikopenfeatures

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	flipt "github.com/open-feature/go-sdk-contrib/providers/flipt/pkg/provider"
	"github.com/open-feature/go-sdk/openfeature"
)

type flag_type string

const (
	int_type    flag_type = "int"
	string_type flag_type = "string"
	float_type  flag_type = "float"
	bool_type   flag_type = "bool"
	object_type flag_type = "object"
)

// Config the plugin configuration.
type Config struct {
	Endpoint            string            `json:"endpoint"`
	Authorization       string            `json:"authorization"`
	Provider            string            `json:"provider"`
	ContextHeaderKeys   []string          `json:"contextHeaderKeys"`
	Service             string            `json:"service"`
	FeatureHeaderPrefix string            `json:"featureHeaderPrefix"`
	Flags               map[string]string `json:"flags"`
	UserHeader          string            `json:"userHeader"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// Example a plugin.
type OpenFeatures struct {
	client              openfeature.Client
	provider            string
	next                http.Handler
	endpoint            string
	name                string
	headers             []string
	featureHeaderPrefix string
	flags               map[string]string
	userheader          string
	// ...
}

// New created a new plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	switch config.Provider {
	case "flipt":
		openfeature.SetProvider(flipt.NewProvider(
			flipt.WithAddress(config.Endpoint),
		))

	default:
		openfeature.SetProvider(openfeature.NoopProvider{})
	}

	featureHeaderPrefix := "openfeature_"
	if config.FeatureHeaderPrefix != "" {
		featureHeaderPrefix = config.FeatureHeaderPrefix
	}

	return &OpenFeatures{
		client:              *openfeature.NewClient(config.Service),
		provider:            config.Provider,
		next:                next,
		headers:             config.ContextHeaderKeys,
		endpoint:            config.Endpoint,
		name:                name,
		featureHeaderPrefix: featureHeaderPrefix,
		flags:               config.Flags,
		userheader:          config.UserHeader,
	}, nil
}

type Flag struct {
	key   string
	value string
}

func (middleware *OpenFeatures) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	evaluateContext := map[string]interface{}{}
	for _, v := range middleware.headers {
		evaluateContext[v] = req.Header.Get(v)
	}
	userId := ""
	if middleware.userheader != "" {
		userId = req.Header.Get(middleware.userheader)
	}

	ch := make(chan Flag, len(middleware.flags))
	var wg sync.WaitGroup

	// TODO: bulk evaluation
	for k, v := range middleware.flags {
		wg.Add(1)
		go func() {
			defer wg.Done()
			flag := Flag{
				key: middleware.featureHeaderPrefix + k,
			}
			switch v {
			case string(int_type):
				{
					value, err := middleware.client.IntValue(req.Context(), k, 0, openfeature.NewEvaluationContext(userId, evaluateContext))
					if err == nil {
						flag.value = strconv.FormatInt(value, 10)
					}
				}

			case string(float_type):
				{
					value, err := middleware.client.FloatValue(req.Context(), k, 0, openfeature.NewEvaluationContext(userId, evaluateContext))
					if err == nil {
						flag.value = strconv.FormatFloat(value, 'f', -1, 64)
					}
				}

			case string(string_type):
				{
					value, err := middleware.client.StringValue(req.Context(), k, "", openfeature.NewEvaluationContext(userId, evaluateContext))
					if err == nil {
						flag.value = value
					}
				}

			case string(bool_type):
				{
					value, err := middleware.client.BooleanValue(req.Context(), k, false, openfeature.NewEvaluationContext(userId, evaluateContext))
					if err == nil {
						flag.value = strconv.FormatBool(value)
					}
				}

			case string(object_type):
				{
					value, err := middleware.client.ObjectValue(req.Context(), k, map[string]interface{}{}, openfeature.NewEvaluationContext(userId, evaluateContext))
					if err == nil {
						b, _ := json.Marshal(value)
						flag.value = string(b)
					}
				}
			}
			ch <- flag
		}()
	}
	wg.Wait()
	close(ch)

	for flag := range ch {
		req.Header.Set(flag.key, flag.value)
	}

	middleware.next.ServeHTTP(rw, req)
}
