package swagger

import (
	_ "embed"
	"net/http"
	"strings"
)

//go:embed openapi.yaml
var OpenAPIYAML []byte

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>LogiFlows API Documentation | Swagger UI</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.18.2/swagger-ui.css">
  <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%236366f1'><path d='M3 3h18v18H3V3zm16 16V5H5v14h14z'/></svg>">
  <style>
    html {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
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
    .logiflows-header {
      background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 100%);
      color: #ffffff;
      padding: 1rem 2rem;
      display: flex;
      align-items: center;
      justify-content: space-between;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    }
    .logiflows-brand {
      display: flex;
      align-items: center;
      gap: 0.75rem;
    }
    .logiflows-brand h1 {
      font-size: 1.25rem;
      font-weight: 700;
      margin: 0;
      letter-spacing: -0.025em;
    }
    .logiflows-brand span {
      background: #4f46e5;
      font-size: 0.75rem;
      padding: 0.2rem 0.5rem;
      border-radius: 9999px;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .logiflows-links {
      display: flex;
      align-items: center;
      gap: 1.25rem;
    }
    .logiflows-links a {
      color: #cbd5e1;
      text-decoration: none;
      font-size: 0.875rem;
      font-weight: 500;
      transition: color 0.15s ease-in-out;
    }
    .logiflows-links a:hover {
      color: #ffffff;
    }
    .swagger-ui .info {
      margin: 2rem 0;
    }
    .swagger-ui .info .title {
      color: #0f172a;
    }
    .swagger-ui .btn.authorize {
      background-color: #4f46e5;
      border-color: #4f46e5;
      color: #fff;
    }
    .swagger-ui .btn.authorize svg {
      fill: #fff;
    }
  </style>
</head>
<body>
  <header class="logiflows-header">
    <div class="logiflows-brand">
      <h1>LogiFlows Authoritative Gateway</h1>
      <span>OpenAPI 3.0</span>
    </div>
    <nav class="logiflows-links">
      <a href="/api/v1/openapi.yaml" target="_blank" download="openapi.yaml">Download OpenAPI Spec</a>
      <a href="/api/v1/healthz" target="_blank">Health Check</a>
      <a href="/" target="_blank">API Root</a>
    </nav>
  </header>

  <div id="swagger-ui"></div>

  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.18.2/swagger-ui-bundle.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.18.2/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/api/v1/openapi.yaml",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        displayRequestDuration: true,
        persistAuthorization: true,
        filter: true,
        tryItOutEnabled: true,
        defaultModelsExpandDepth: 1,
        defaultModelExpandDepth: 1
      });
    };
  </script>
</body>
</html>
`

// SpecHandler serves the raw OpenAPI 3.0 specification in YAML format
func SpecHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(OpenAPIYAML)
}

// UIHandler serves the interactive Swagger UI web application
func UIHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "openapi.yaml") {
		SpecHandler(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}
