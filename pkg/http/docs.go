// Package http provides Swagger documentation handler for the service.
package http

import (
	"embed"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

// DocsConfig holds the configuration for the documentation server.
type DocsConfig struct {
	ServiceName string
	Version     string
	SpecPath    string // Path to openapi.yaml
}

// swaggerUITemplate is the HTML template for Swagger UI
const swaggerUITemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.ServiceName}} API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
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
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #1a1a2e;
        }
        .swagger-ui .topbar .download-url-wrapper .download-url-button {
            background: #4a3f69;
        }
        .swagger-ui .info .title {
            color: #1a1a2e;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "{{.SpecURL}}",
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
                validatorUrl: null,
                docExpansion: "list",
                defaultModelsExpandDepth: 1,
                displayRequestDuration: true
            });
        };
    </script>
</body>
</html>
`

// SetupDocsRoutes configures documentation routes on the given mux.
// Routes created:
//   - GET /docs - Swagger UI
//   - GET /docs/openapi.yaml - OpenAPI specification
func SetupDocsRoutes(mux *http.ServeMux, config DocsConfig) {
	// Template for Swagger UI
	tmpl := template.Must(template.New("swagger").Parse(swaggerUITemplate))

	// Serve Swagger UI
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/docs" && r.URL.Path != "/docs/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := struct {
			ServiceName string
			SpecURL     string
		}{
			ServiceName: config.ServiceName,
			SpecURL:     "/docs/openapi.yaml",
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render documentation", http.StatusInternalServerError)
		}
	})

	// Serve OpenAPI spec
	mux.HandleFunc("GET /docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		specPath := config.SpecPath
		if specPath == "" {
			// Default path relative to executable
			execPath, _ := os.Executable()
			specPath = filepath.Join(filepath.Dir(execPath), "docs", "openapi.yaml")
		}

		content, err := os.ReadFile(specPath)
		if err != nil {
			// Try embedded spec or fallback
			http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/yaml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(content)
	})
}

// SetupDocsRoutesWithEmbed configures documentation routes using embedded files.
func SetupDocsRoutesWithEmbed(mux *http.ServeMux, config DocsConfig, specFS embed.FS, specFile string) {
	// Template for Swagger UI
	tmpl := template.Must(template.New("swagger").Parse(swaggerUITemplate))

	// Serve Swagger UI
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/docs" && r.URL.Path != "/docs/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := struct {
			ServiceName string
			SpecURL     string
		}{
			ServiceName: config.ServiceName,
			SpecURL:     "/docs/openapi.yaml",
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render documentation", http.StatusInternalServerError)
		}
	})

	// Serve OpenAPI spec from embedded filesystem
	mux.HandleFunc("GET /docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		content, err := specFS.ReadFile(specFile)
		if err != nil {
			http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/yaml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(content)
	})
}
