// Package docs містить OpenAPI-специфікацію API, вбудовану в бінарник.
package docs

import _ "embed"

// Spec — вміст файлу openapi.yaml. Директива go:embed читає файл під час
// збірки, тому окремий файл поруч із програмою не потрібен.
//
//go:embed openapi.yaml
var Spec []byte
