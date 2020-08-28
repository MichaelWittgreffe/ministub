package config

// Request represents a config definition for a request to make
type Request struct {
	URL              string                 `yaml:"url"`
	Headers          map[string]string      `yaml:"headers"`
	Body             map[string]interface{} `yaml:"body"`
	ExpectedResponse *Response              `yaml:"expectedResponse"`
}
