package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"homework/internal/docs"
)

// swaggerPage — сторінка Swagger UI. Саму бібліотеку браузер завантажує з CDN,
// а специфікацію бере з нашого /openapi.yaml.
const swaggerPage = `<!doctype html>
<html lang="uk">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Books API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({ url: "/openapi.yaml", dom_id: "#swagger-ui" });
  </script>
</body>
</html>
`

// OpenAPI відповідає на GET /openapi.yaml: віддає специфікацію API.
func OpenAPI(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", docs.Spec)
}

// SwaggerUI відповідає на GET /swagger: інтерактивна документація в браузері.
func SwaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerPage))
}
