package swagger

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var openAPISpec []byte

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Research Helper API Documentation</title>
  <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.18.2/swagger-ui.min.css" />
  <link rel="icon" type="image/png" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.18.2/favicon-32x32.png" sizes="32x32" />
  <style>
    html {
      box-sizing: border-box;
      overflow-y: scroll;
    }
    *, *:before, *:after {
      box-sizing: inherit;
    }
    body {
      margin: 0;
      background: #f8fafc;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    }
    .topbar {
      display: none !important;
    }
    .swagger-ui {
      max-width: 1240px;
      margin: 0 auto;
      padding: 24px;
    }
    .swagger-ui .info .title {
      font-size: 2.2rem;
      font-weight: 700;
      color: #0f172a;
    }
    .swagger-ui .scheme-container {
      background: #ffffff;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      padding: 12px 20px;
      margin-bottom: 24px;
      box-shadow: 0 1px 3px rgba(0,0,0,0.05);
    }
    .swagger-ui .opblock {
      border-radius: 8px;
      box-shadow: 0 1px 2px rgba(0,0,0,0.04);
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.18.2/swagger-ui-bundle.min.js" charset="UTF-8"></script>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.18.2/swagger-ui-standalone-preset.min.js" charset="UTF-8"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/api/v1/swagger.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout",
        docExpansion: "list",
        filter: true,
        tryItOutEnabled: true,
        displayRequestDuration: true
      });
    };
  </script>
</body>
</html>`

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ServeSpec returns the embedded OpenAPI 3.0 JSON specification.
func (h *Handler) ServeSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}

// ServeUI returns the Swagger UI interactive documentation page.
func (h *Handler) ServeUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}
