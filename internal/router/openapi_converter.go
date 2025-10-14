package router

import (
	"encoding/json"
)

// OpenAPI3Spec структура для правильного порядка полей
type OpenAPI3Spec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       interface{}            `json:"info"`
	Servers    []interface{}          `json:"servers,omitempty"`
	Tags       []interface{}          `json:"tags,omitempty"`
	Paths      map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components,omitempty"`
	Security   []interface{}          `json:"security,omitempty"`
}

// GetOpenAPI3Spec конвертирует Swagger 2.0 в OpenAPI 3.1
func GetOpenAPI3Spec(swagger2JSON []byte) ([]byte, error) {
	var swagger2 map[string]interface{}
	if err := json.Unmarshal(swagger2JSON, &swagger2); err != nil {
		return nil, err
	}

	// Создаём структуру OpenAPI 3.0.3 (не 3.1.0!)
	spec := OpenAPI3Spec{
		OpenAPI: "3.0.3",
	}

	// Info
	if info, ok := swagger2["info"]; ok {
		spec.Info = info
	}

	// Servers
	servers := []interface{}{}
	host := ""
	basePath := ""
	scheme := "http"

	if h, ok := swagger2["host"].(string); ok {
		host = h
	}
	if bp, ok := swagger2["basePath"].(string); ok {
		basePath = bp
	}
	if schemes, ok := swagger2["schemes"].([]interface{}); ok && len(schemes) > 0 {
		if s, ok := schemes[0].(string); ok {
			scheme = s
		}
	}

	if host != "" {
		serverURL := scheme + "://" + host + basePath
		servers = append(servers, map[string]interface{}{
			"url": serverURL,
		})
		spec.Servers = servers
	}

	// Tags
	if tags, ok := swagger2["tags"]; ok {
		if tagsList, ok := tags.([]interface{}); ok {
			spec.Tags = tagsList
		}
	}

	// Paths - конвертируем и заменяем $ref
	if paths, ok := swagger2["paths"]; ok {
		if pathsMap, ok := paths.(map[string]interface{}); ok {
			convertedPaths := convertPaths(pathsMap)
			spec.Paths = convertRefs(convertedPaths).(map[string]interface{})
		}
	}

	// Components
	components := make(map[string]interface{})
	if definitions, ok := swagger2["definitions"]; ok {
		// Конвертируем $ref внутри schemas
		components["schemas"] = convertRefs(definitions)
	}
	if securityDefs, ok := swagger2["securityDefinitions"]; ok {
		if secDefsMap, ok := securityDefs.(map[string]interface{}); ok {
			components["securitySchemes"] = convertSecurityDefinitions(secDefsMap)
		}
	}
	if len(components) > 0 {
		spec.Components = components
	}

	// Security
	if security, ok := swagger2["security"]; ok {
		if secList, ok := security.([]interface{}); ok {
			spec.Security = secList
		}
	}

	return json.MarshalIndent(spec, "", "  ")
}

// convertPaths converts Swagger 2.0 paths to OpenAPI 3.1 format
func convertPaths(paths map[string]interface{}) map[string]interface{} {
	converted := make(map[string]interface{})

	for path, pathItem := range paths {
		if pathItemMap, ok := pathItem.(map[string]interface{}); ok {
			convertedPathItem := make(map[string]interface{})

			for method, operation := range pathItemMap {
				if operationMap, ok := operation.(map[string]interface{}); ok {
					convertedPathItem[method] = convertOperation(operationMap)
				} else {
					convertedPathItem[method] = operation
				}
			}

			converted[path] = convertedPathItem
		}
	}

	return converted
}

// convertOperation converts a single operation from Swagger 2.0 to OpenAPI 3.1
func convertOperation(operation map[string]interface{}) map[string]interface{} {
	converted := make(map[string]interface{})

	// Copy most fields as is
	for key, value := range operation {
		if key == "parameters" {
			if params, ok := value.([]interface{}); ok {
				converted["parameters"] = convertParameters(params)
			}
		} else if key == "responses" {
			if responses, ok := value.(map[string]interface{}); ok {
				converted["responses"] = convertResponses(responses)
			}
		} else if key == "consumes" || key == "produces" {
			// These are moved to requestBody and responses in OpenAPI 3.x
			continue
		} else {
			converted[key] = value
		}
	}

	// Handle consumes/produces -> requestBody
	if params, ok := operation["parameters"].([]interface{}); ok {
		if requestBody := extractRequestBody(params, operation); requestBody != nil {
			converted["requestBody"] = requestBody
			// Remove body parameters from parameters list
			converted["parameters"] = filterNonBodyParameters(params)
		}
	}

	return converted
}

// convertParameters converts parameters array
func convertParameters(params []interface{}) []interface{} {
	converted := []interface{}{}

	for _, param := range params {
		if paramMap, ok := param.(map[string]interface{}); ok {
			// Skip body parameters as they become requestBody in OpenAPI 3.x
			if in, ok := paramMap["in"].(string); ok && in == "body" {
				continue
			}

			// Конвертируем параметры в OpenAPI 3.0 формат
			convertedParam := make(map[string]interface{})

			// Копируем основные поля
			if name, ok := paramMap["name"]; ok {
				convertedParam["name"] = name
			}
			if in, ok := paramMap["in"]; ok {
				convertedParam["in"] = in
			}
			if description, ok := paramMap["description"]; ok {
				convertedParam["description"] = description
			}
			if required, ok := paramMap["required"]; ok {
				convertedParam["required"] = required
			}

			// Создаём schema для type, format, default и т.д.
			schema := make(map[string]interface{})

			if paramType, ok := paramMap["type"]; ok {
				schema["type"] = paramType
			}
			if format, ok := paramMap["format"]; ok {
				schema["format"] = format
			}
			if defaultVal, ok := paramMap["default"]; ok {
				schema["default"] = defaultVal
			}
			if minLength, ok := paramMap["minLength"]; ok {
				schema["minLength"] = minLength
			}
			if maxLength, ok := paramMap["maxLength"]; ok {
				schema["maxLength"] = maxLength
			}
			if minimum, ok := paramMap["minimum"]; ok {
				schema["minimum"] = minimum
			}
			if maximum, ok := paramMap["maximum"]; ok {
				schema["maximum"] = maximum
			}
			if enum, ok := paramMap["enum"]; ok {
				schema["enum"] = enum
			}

			if len(schema) > 0 {
				convertedParam["schema"] = schema
			}

			converted = append(converted, convertedParam)
		}
	}

	return converted
}

// filterNonBodyParameters removes body parameters from the list
func filterNonBodyParameters(params []interface{}) []interface{} {
	filtered := []interface{}{}

	for _, param := range params {
		if paramMap, ok := param.(map[string]interface{}); ok {
			if in, ok := paramMap["in"].(string); ok && in != "body" {
				filtered = append(filtered, param)
			}
		}
	}

	return filtered
}

// extractRequestBody extracts body parameter and converts to requestBody
func extractRequestBody(params []interface{}, operation map[string]interface{}) map[string]interface{} {
	for _, param := range params {
		if paramMap, ok := param.(map[string]interface{}); ok {
			if in, ok := paramMap["in"].(string); ok && in == "body" {
				requestBody := make(map[string]interface{})

				if required, ok := paramMap["required"].(bool); ok {
					requestBody["required"] = required
				}

				if description, ok := paramMap["description"].(string); ok {
					requestBody["description"] = description
				}

				// Get content type from consumes
				contentType := "application/json"
				if consumes, ok := operation["consumes"].([]interface{}); ok && len(consumes) > 0 {
					if ct, ok := consumes[0].(string); ok {
						contentType = ct
					}
				}

				content := make(map[string]interface{})
				mediaType := make(map[string]interface{})

				if schema, ok := paramMap["schema"]; ok {
					mediaType["schema"] = schema
				}

				content[contentType] = mediaType
				requestBody["content"] = content

				return requestBody
			}
		}
	}

	return nil
}

// convertResponses converts responses object
func convertResponses(responses map[string]interface{}) map[string]interface{} {
	converted := make(map[string]interface{})

	for code, response := range responses {
		if responseMap, ok := response.(map[string]interface{}); ok {
			convertedResponse := make(map[string]interface{})

			// Copy description
			if description, ok := responseMap["description"]; ok {
				convertedResponse["description"] = description
			}

			// Convert schema to content
			if schema, ok := responseMap["schema"]; ok {
				content := make(map[string]interface{})
				mediaType := make(map[string]interface{})
				mediaType["schema"] = schema
				content["application/json"] = mediaType
				convertedResponse["content"] = content
			}

			// Copy headers if present
			if headers, ok := responseMap["headers"]; ok {
				convertedResponse["headers"] = headers
			}

			converted[code] = convertedResponse
		}
	}

	return converted
}

// convertSecurityDefinitions converts securityDefinitions to securitySchemes
func convertSecurityDefinitions(securityDefs map[string]interface{}) map[string]interface{} {
	converted := make(map[string]interface{})

	for name, secDef := range securityDefs {
		if secDefMap, ok := secDef.(map[string]interface{}); ok {
			convertedSecDef := make(map[string]interface{})

			// Type conversion
			if secType, ok := secDefMap["type"].(string); ok {
				if secType == "basic" {
					convertedSecDef["type"] = "http"
					convertedSecDef["scheme"] = "basic"
				} else if secType == "apiKey" {
					convertedSecDef["type"] = "apiKey"
					if in, ok := secDefMap["in"]; ok {
						convertedSecDef["in"] = in
					}
					if keyName, ok := secDefMap["name"]; ok {
						convertedSecDef["name"] = keyName
					}
				} else if secType == "oauth2" {
					convertedSecDef["type"] = "oauth2"
					// OAuth2 flows need more complex conversion
					if flow, ok := secDefMap["flow"].(string); ok {
						flows := make(map[string]interface{})
						flowObj := make(map[string]interface{})

						if authURL, ok := secDefMap["authorizationUrl"]; ok {
							flowObj["authorizationUrl"] = authURL
						}
						if tokenURL, ok := secDefMap["tokenUrl"]; ok {
							flowObj["tokenUrl"] = tokenURL
						}
						if scopes, ok := secDefMap["scopes"]; ok {
							flowObj["scopes"] = scopes
						}

						flows[flow] = flowObj
						convertedSecDef["flows"] = flows
					}
				}
			}

			// Copy description
			if description, ok := secDefMap["description"]; ok {
				convertedSecDef["description"] = description
			}

			converted[name] = convertedSecDef
		}
	}

	return converted
}

func convertRefs(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			if key == "$ref" {
				if refStr, ok := value.(string); ok {
					if len(refStr) > 14 && refStr[:14] == "#/definitions/" {
						result[key] = "#/components/schemas/" + refStr[14:]
					} else {
						result[key] = value
					}
				} else {
					result[key] = value
				}
			} else {
				result[key] = convertRefs(value)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertRefs(item)
		}
		return result
	default:
		return v
	}
}
