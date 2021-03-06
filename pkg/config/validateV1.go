package config

import (
	"fmt"
)

// validateV1Config validates an incoming config against version 1
func validateV1Config(cfg *Config) error {
	serviceNames := make(map[string]bool)
	if len(cfg.Services) > 0 {
		for serviceName, entry := range cfg.Services {
			if len(entry.Hostname) > 0 {
				var err error
				entry.Hostname, err = getEnvValueForField(entry.Hostname)
				if err != nil {
					return err
				}
			}

			if len(entry.Hostname) == 0 || (entry.Port <= 0 || entry.Port > 65535) {
				return fmt.Errorf("Invalid Service Entry For Service: %s", serviceName)
			}
			serviceNames[serviceName] = true
		}
	}

	if len(cfg.StartupActions) > 0 {
		if err := validateV1Actions(cfg.StartupActions, serviceNames, cfg.Requests); err != nil {
			return fmt.Errorf("Failed Validating Startup Actions: %s", err.Error())
		}
	}

	if len(cfg.Requests) > 0 {
		for reqName, entry := range cfg.Requests {
			if err := validateV1Request(reqName, entry); err != nil {
				return err
			}
		}
	}

	if len(cfg.Endpoints) > 0 {
		for url, methodMap := range cfg.Endpoints {
			for method, entry := range methodMap {
				if err := validateV1Endpoint(url, method, entry, serviceNames, cfg.Requests); err != nil {
					return err
				}
			}
		}
	} else {
		return fmt.Errorf("No Endpoints Set")
	}

	return nil
}

// validateV1Actions ensures an 'actions' field for a V1 config is correct
func validateV1Actions(actions []map[string]interface{}, serviceNames map[string]bool, requests map[string]*Request) error {
	for _, actionMap := range actions {
		for actionName, actionEntry := range actionMap {
			if !supportedAction(actionName) {
				return fmt.Errorf("Action Not Supported: %s", actionName)
			}
			if actionName == "request" {
				if newActionEntry, valid := actionEntry.(map[interface{}]interface{}); valid {
					if target, found := newActionEntry["target"]; found {
						if _, found = serviceNames[target.(string)]; !found {
							return fmt.Errorf("Service Not Defined For Request Action: %s", target)
						}
					} else {
						return fmt.Errorf("No Target Defined For Request Action")
					}

					if requestName, found := newActionEntry["id"]; found {
						if _, found := requests[requestName.(string)]; !found {
							return fmt.Errorf("Request ID Not Defined For Request Action %s", requestName)
						}
					} else {
						return fmt.Errorf("No Request ID Defined For Request Action")
					}
				} else {
					return fmt.Errorf("Invalid Request Value For Request Action")
				}
			} else if actionName == "delay" {
				if _, valid := actionEntry.(int); !valid {
					return fmt.Errorf("Invalid Delay Value For Request Action")
				}
			}
		}
	}

	return nil
}

// validateV1Parameters ensures the given parameter set is valid according to v1 schema
func validateV1Parameters(params map[string]*ParamEntry) error {
	for field, paramEntry := range params {
		if !supportedType(paramEntry.Type) {
			return fmt.Errorf("Field %s Type Is Not Supported: %s", field, paramEntry.Type)
		}
	}
	return nil
}

// validateV1Method checks if the given http method is valid
func validateV1Method(method string) bool {
	switch {
	case method == "get":
		return true
	case method == "post":
		return true
	case method == "put":
		return true
	case method == "delete":
		return true
	default:
		return false
	}
}

// validateV1Protocol checks if the given protocol is valid
func validateV1Protocol(protocol string) bool {
	switch {
	case protocol == "http":
		return true
	case protocol == "https":
		return true
	default:
		return false
	}
}

// validateV1Endpoints ensures a given endpoint definition is valid
func validateV1Endpoint(url, method string, entry *Endpoint, serviceNames map[string]bool, requests map[string]*Request) error {
	if entry.Params != nil {
		if entry.Params.Path != nil && len(entry.Params.Path) > 0 {
			if err := validateV1Parameters(entry.Params.Path); err != nil {
				return fmt.Errorf("Path Param For URL %s, Method %s Not Valid: %s", url, method, err.Error())
			}
		}
		if entry.Params.Query != nil && len(entry.Params.Query) > 0 {
			if err := validateV1Parameters(entry.Params.Query); err != nil {
				return fmt.Errorf("Query Param For URL %s, Method %s Not Valid: %s", url, method, err.Error())
			}
		}
	}

	if entry.Recieves != nil {
		for name, exType := range entry.Recieves.Body {
			if !supportedType(exType) {
				return fmt.Errorf("Body Field Type For URL %s, Method %s, Field %s Type Is Not Supported: %s", url, method, name, exType)
			}
		}
	}

	if entry.Responses == nil && entry.Response == 0 {
		return fmt.Errorf("Response Not Set For URL %s, Method %s", url, method)
	}

	if entry.Responses != nil {
		totalWeight := 0
		for _, respEntry := range entry.Responses {
			totalWeight += respEntry.Weight

			if len(respEntry.Actions) > 0 {
				if err := validateV1Actions(respEntry.Actions, serviceNames, requests); err != nil {
					return fmt.Errorf("Error Validating Response Action URL %s, Method %s: %s", url, method, err.Error())
				}
			}
		}
		if totalWeight != 100 {
			return fmt.Errorf("Response Weighting For URL %s, Method %s, Does Not Equal 100", url, method)
		}
	}

	if entry.Actions != nil && len(entry.Actions) > 0 {
		if err := validateV1Actions(entry.Actions, serviceNames, requests); err != nil {
			return fmt.Errorf("Error Validating URL %s, Method %s: %s", url, method, err.Error())
		}
	}

	return nil
}

// validateV1Request ensures a given request field is valid; only mandatory fields are URL and expected response code
func validateV1Request(reqName string, entry *Request) error {
	if entry == nil {
		return fmt.Errorf("Entry Is Nil")
	}

	if len(reqName) == 0 {
		return fmt.Errorf("reqName Is Empty")
	}

	if len(entry.URL) == 0 {
		return fmt.Errorf("URL For Request %s Is Empty", reqName)
	}

	if !validateV1Method(entry.Method) {
		return fmt.Errorf("Method %s For Request %s Not Supported", entry.Method, reqName)
	}

	if len(entry.Protocol) > 0 {
		if !validateV1Protocol(entry.Protocol) {
			return fmt.Errorf("Protocol %s For Request %s Not Supported", entry.Protocol, reqName)
		}
	} else {
		entry.Protocol = "http" // default if ommitted
	}

	if entry.ExpectedResponse != nil {
		if entry.ExpectedResponse.StatusCode == 0 {
			return fmt.Errorf("Status Code For Request %s Is Invalid", reqName)
		}

		if entry.ExpectedResponse.Body != nil && len(entry.ExpectedResponse.Body) > 0 {
			for expectedField, expectedType := range entry.ExpectedResponse.Body {
				if strExpectedType, ok := expectedType.(string); ok && !supportedType(strExpectedType) {
					return fmt.Errorf("Request %s Expected Response Invalid Expected Type For Field %s: %s", reqName, expectedField, expectedType)
				}
			}
		}
	}

	if entry.Body != nil && len(entry.Body) > 0 {
		entry.Body = validateJSON(entry.Body).(map[string]interface{})
	}

	return nil
}
