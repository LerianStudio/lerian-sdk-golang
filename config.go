package lerian

import (
	"net/http"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"
	"github.com/LerianStudio/lerian-sdk-golang/reporter"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
)

// Config defines the explicit construction contract for the root Lerian SDK
// client. A product is enabled when its config pointer is non-nil.
type Config struct {
	Midaz    *midaz.Config
	Matcher  *matcher.Config
	Tracer   *tracer.Config
	Reporter *reporter.Config
	Fees     *fees.Config

	Debug         bool
	HTTPClient    *http.Client
	RetryConfig   *retry.Config
	Observability ObservabilityConfig
}

// ObservabilityConfig controls the OTel provider created by the root client.
// When all pillars are disabled, the client uses noop observability.
type ObservabilityConfig struct {
	Traces            bool
	Metrics           bool
	Logs              bool
	CollectorEndpoint string
}

// LoadConfigFromEnv builds a root [Config] from supported `LERIAN_*`
// environment variables. Products are enabled only when at least one
// environment variable for that product is present.
func LoadConfigFromEnv() Config {
	return Config{
		Midaz:    loadMidazConfigFromEnv(),
		Matcher:  loadMatcherConfigFromEnv(),
		Tracer:   loadTracerConfigFromEnv(),
		Reporter: loadReporterConfigFromEnv(),
		Fees:     loadFeesConfigFromEnv(),
		Debug:    envBool(envDebug),
	}
}
