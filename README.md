# frp-schemas

JSON Schemas for frpc/frps configuration files with IDE autocompletion and validation support.

## Available Schemas

- **frpc**: `https://github.com/LazuliKao/frp-schemas/raw/refs/heads/main/frpc-schema.json`
- **frps**: `https://github.com/LazuliKao/frp-schemas/raw/refs/heads/main/frps-schema.json`

## Usage

### YAML Configuration

Add the `# yaml-language-server` comment at the top of your configuration file:

```yaml
# yaml-language-server: $schema=https://github.com/LazuliKao/frp-schemas/raw/refs/heads/main/frpc-schema.json

serverAddr: "0.0.0.0"
serverPort: 7000

proxies:
  - name: ssh
    type: tcp
    localIP: 127.0.0.1
    localPort: 22
    remotePort: 6000
```

**Supported editors**: VS Code (with [YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)), IntelliJ IDEA, WebStorm, and other editors with YAML Language Server support.

### JSON Configuration

Add the `$schema` property at the root level:

```json
{
  "$schema": "https://github.com/LazuliKao/frp-schemas/raw/refs/heads/main/frpc-schema.json",
  "serverAddr": "0.0.0.0",
  "serverPort": 7000,
  "proxies": [
    {
      "name": "ssh",
      "type": "tcp",
      "localIP": "127.0.0.1",
      "localPort": 22,
      "remotePort": 6000
    }
  ]
}
```

**Supported editors**: VS Code, IntelliJ IDEA, WebStorm, Sublime Text, and most modern editors with built-in JSON schema support.

### TOML Configuration

Add the `$schema` key at the top level:

```toml
"$schema" = "https://github.com/LazuliKao/frp-schemas/raw/refs/heads/main/frpc-schema.json"

serverAddr = "0.0.0.0"
serverPort = 7000

[[proxies]]
name = "ssh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = 6000
```

**Note**: Schema validation support for TOML is limited. Some editors may support it through extensions:
- VS Code: [Even Better TOML](https://marketplace.visualstudio.com/items?itemName=tamasfe.even-better-toml)
- IntelliJ IDEA: Built-in TOML plugin

### INI Configuration (frp legacy format)

INI format does not support schema validation. Consider migrating to YAML, JSON, or TOML for better IDE support.

## Benefits

- **Autocompletion**: Get suggestions for configuration keys as you type
- **Validation**: Catch configuration errors before runtime
- **Documentation**: Inline documentation for all configuration options
- **Type checking**: Ensure correct value types for each setting

## Editor Setup

### VS Code

1. Install the [YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)
2. The schema will be automatically detected from the `# yaml-language-server` comment

### IntelliJ IDEA / WebStorm

Schema references are automatically recognized in YAML and JSON files. No additional setup required.

### Other Editors

Consult your editor's documentation for JSON Schema and YAML Language Server support.

## Contributing

Issues and pull requests are welcome at [https://github.com/LazuliKao/frp-schemas](https://github.com/LazuliKao/frp-schemas).

## License

This project provides JSON schemas for the [frp](https://github.com/fatedier/frp) project configuration files.