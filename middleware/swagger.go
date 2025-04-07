package middleware

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// Swagger serves the Swagger UI and the OpenAPI specification
func Swagger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/api-docs" || c.Request.URL.Path == "/api-docs/" {
			c.Redirect(http.StatusMovedPermanently, "/api-docs/index.html")
			return
		}

		if c.Request.URL.Path == "/api-docs/swagger.yaml" {
			// Serve the swagger.yaml file
			c.File("./swagger.yaml")
			return
		}

		if filepath.Ext(c.Request.URL.Path) == "" || c.Request.URL.Path == "/api-docs/index.html" {
			// Serve the Swagger UI HTML
			serveSwaggerUI(c)
			return
		}

		// Serve static assets for Swagger UI
		workDir, _ := os.Getwd()
		filesDir := filepath.Join(workDir, "public", "swagger")
		relativePath := c.Request.URL.Path[9:] // Remove "/api-docs/" prefix
		fileToServe := filepath.Join(filesDir, relativePath)

		if _, err := os.Stat(fileToServe); os.IsNotExist(err) {
			c.Status(http.StatusNotFound)
			return
		}

		c.File(fileToServe)
	}
}

// serveSwaggerUI serves the Swagger UI HTML page
func serveSwaggerUI(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4.5.0/swagger-ui.css" >
  <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4.5.0/favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4.5.0/favicon-16x16.png" sizes="16x16" />
  <style>
    html {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }
    *,
    *:before,
    *:after {
      box-sizing: inherit;
    }
    body {
      margin:0;
      background: #fafafa;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4.5.0/swagger-ui-bundle.js" charset="UTF-8"> </script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4.5.0/swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
  <script>
  window.onload = function() {
    const ui = SwaggerUIBundle({
      url: "/api-docs/swagger.yaml",
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIStandalonePreset
      ],
      plugins: [
        SwaggerUIBundle.plugins.DownloadUrl
      ],
      layout: "StandaloneLayout"
    });
    window.ui = ui;
  }
  </script>
</body>
</html>
`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
