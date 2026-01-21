package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/invopop/jsonschema"
)

func main() {
	// Generate schema for frpc (client) configuration
	clientSchema := generateClientSchema()
	clientJSON, err := json.MarshalIndent(clientSchema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating client schema: %v\n", err)
		os.Exit(1)
	}

	// Post-process the JSON to enhance proxy schemas
	var clientSchemaMap map[string]any
	if err := json.Unmarshal(clientJSON, &clientSchemaMap); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling client schema: %v\n", err)
		os.Exit(1)
	}

	enhanceProxySchemasJSON(clientSchemaMap)

	clientJSON, err = json.MarshalIndent(clientSchemaMap, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling enhanced client schema: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile("frpc-schema.json", clientJSON, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing client schema file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated frpc-schema.json")

	// Generate schema for frps (server) configuration
	serverSchema := generateServerSchema()
	serverJSON, err := json.MarshalIndent(serverSchema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating server schema: %v\n", err)
		os.Exit(1)
	}

	// Make server version optional
	var serverSchemaMap map[string]any
	if err := json.Unmarshal(serverJSON, &serverSchemaMap); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling server schema: %v\n", err)
		os.Exit(1)
	}
	makeVersionOptional(serverSchemaMap, "ServerConfig")

	serverJSON, err = json.MarshalIndent(serverSchemaMap, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling enhanced server schema: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile("frps-schema.json", serverJSON, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing server schema file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated frps-schema.json")

	fmt.Println("\nJSON schemas generated successfully!")
	fmt.Printf("  - frpc-schema.json (client configuration)\n")
	fmt.Printf("  - frps-schema.json (server configuration)\n")
}

func generateClientSchema() *jsonschema.Schema {
	schema := jsonschema.Reflect(&v1.ClientConfig{})
	schema.Title = "FRP Client Configuration"
	schema.Description = "JSON Schema for frpc (FRP client) configuration file"

	return schema
}

func generateServerSchema() *jsonschema.Schema {
	schema := jsonschema.Reflect(&v1.ServerConfig{})
	schema.Title = "FRP Server Configuration"
	schema.Description = "JSON Schema for frps (FRP server) configuration file"

	return schema
}

func enhanceProxySchemasJSON(schemaMap map[string]any) {
	defs, ok := schemaMap["$defs"].(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Warning: No $defs found in schema\n")
		return
	}

	// 生成各个具体proxy类型的schema
	proxyTypes := []struct {
		name       string
		typeValue  string
		configType any
	}{
		{"TCPProxyConfig", "tcp", &v1.TCPProxyConfig{}},
		{"UDPProxyConfig", "udp", &v1.UDPProxyConfig{}},
		{"HTTPProxyConfig", "http", &v1.HTTPProxyConfig{}},
		{"HTTPSProxyConfig", "https", &v1.HTTPSProxyConfig{}},
		{"TCPMuxProxyConfig", "tcpmux", &v1.TCPMuxProxyConfig{}},
		{"STCPProxyConfig", "stcp", &v1.STCPProxyConfig{}},
		{"XTCPProxyConfig", "xtcp", &v1.XTCPProxyConfig{}},
		{"SUDPProxyConfig", "sudp", &v1.SUDPProxyConfig{}},
	}

	oneOfSchemas := []map[string]any{}

	for _, pt := range proxyTypes {
		proxySchema := jsonschema.Reflect(pt.configType)
		proxySchemaJSON, _ := json.Marshal(proxySchema)
		var fullSchemaMap map[string]any
		json.Unmarshal(proxySchemaJSON, &fullSchemaMap)

		// 提取实际的schema定义（从$ref引用的定义）
		var proxySchemaMap map[string]any
		if _, ok := fullSchemaMap["$ref"].(string); ok {
			// 从$defs中提取对应的定义
			if proxyDefs, ok := fullSchemaMap["$defs"].(map[string]any); ok {
				// $ref格式为 "#/$defs/HTTPProxyConfig"，提取最后一部分
				defName := pt.name
				if def, ok := proxyDefs[defName].(map[string]any); ok {
					proxySchemaMap = def

					// 设置type字段的约束
					if props, ok := proxySchemaMap["properties"].(map[string]any); ok {
						if typeField, ok := props["type"].(map[string]any); ok {
							typeField["enum"] = []string{pt.typeValue}
							typeField["description"] = "Proxy type (must be " + pt.typeValue + ")"
						}
					}

					// 合并子定义到主schema的$defs中
					for defKey, defValue := range proxyDefs {
						if defKey != pt.name { // 避免重复添加自己
							// 添加前缀避免命名冲突，并更新内部引用
							prefixedKey := pt.name + "_" + defKey
							if defMap, ok := defValue.(map[string]any); ok {
								updateRefs(defMap, "#/$defs/", "#/$defs/"+pt.name+"_")
								defs[prefixedKey] = defMap
							} else {
								defs[prefixedKey] = defValue
							}
						}
					}

					// 更新主schema的内部引用
					updateRefs(proxySchemaMap, "#/$defs/", "#/$defs/"+pt.name+"_")
				}
			}
		} else {
			proxySchemaMap = fullSchemaMap
		}

		// 添加到defs
		defs[pt.name] = proxySchemaMap

		// 添加到oneOf列表
		oneOfSchemas = append(oneOfSchemas, map[string]any{
			"$ref": "#/$defs/" + pt.name,
		})
	}

	// 创建ProxyConfig的oneOf schema
	defs["ProxyConfig"] = map[string]any{
		"title":       "Proxy Configuration",
		"description": "Configuration for a single proxy. The available fields depend on the proxy type.",
		"oneOf":       oneOfSchemas,
	}

	// 更新ClientConfig中的proxies字段
	if clientConfigDef, ok := defs["ClientConfig"].(map[string]any); ok {
		removeRequiredField(clientConfigDef, "version")
		if props, ok := clientConfigDef["properties"].(map[string]any); ok {
			if proxiesProp, ok := props["proxies"].(map[string]any); ok {
				proxiesProp["items"] = map[string]any{
					"$ref": "#/$defs/ProxyConfig",
				}
			}
		}
	}

	// 删除原来的TypedProxyConfig定义
	delete(defs, "TypedProxyConfig")
}

// updateRefs递归更新schema中的$ref引用
func updateRefs(schema map[string]any, oldPrefix, newPrefix string) {
	for key, value := range schema {
		if key == "$ref" {
			if refStr, ok := value.(string); ok {
				if strings.HasPrefix(refStr, oldPrefix) {
					schema[key] = newPrefix + strings.TrimPrefix(refStr, oldPrefix)
				}
			}
		} else if subMap, ok := value.(map[string]any); ok {
			updateRefs(subMap, oldPrefix, newPrefix)
		} else if subArray, ok := value.([]any); ok {
			for _, item := range subArray {
				if itemMap, ok := item.(map[string]any); ok {
					updateRefs(itemMap, oldPrefix, newPrefix)
				}
			}
		}
	}
}

// removeRequiredField removes a field from the required list if present.
func removeRequiredField(def map[string]any, field string) {
	req, ok := def["required"].([]any)
	if !ok {
		return
	}
	filtered := make([]any, 0, len(req))
	for _, v := range req {
		if s, ok := v.(string); ok && s == field {
			continue
		}
		filtered = append(filtered, v)
	}
	def["required"] = filtered
}

// makeVersionOptional removes version from required list for given root def.
func makeVersionOptional(schemaMap map[string]any, defName string) {
	defs, ok := schemaMap["$defs"].(map[string]any)
	if !ok {
		return
	}
	if def, ok := defs[defName].(map[string]any); ok {
		removeRequiredField(def, "version")
	}
}
