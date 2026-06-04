package admin

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"api-service/utils"
)

// AppDisplay extends models.App with token counts for rendering
type AppDisplay struct {
	AppID      string
	Name       string
	Version    string
	IsActive   bool
	TokenCount int
}

// BaseHandler holds shared infrastructure (compiled templates) used by
// all admin domain handlers via embedding.
type BaseHandler struct {
	templates map[string]*template.Template
}

// NewBaseHandler compiles layout + view templates on startup and
// returns a ready-to-use BaseHandler.
func NewBaseHandler(embeddedFS embed.FS) *BaseHandler {
	h := &BaseHandler{
		templates: make(map[string]*template.Template),
	}
	h.initTemplates(embeddedFS)
	return h
}

// initTemplates reads embedded asset directories and builds template compilations
func (h *BaseHandler) initTemplates(embeddedFS embed.FS) {
	subFS, err := fs.Sub(embeddedFS, "web")
	if err != nil {
		panic("failed to map embedded web assets: " + err.Error())
	}

	views := []string{"login", "dashboard", "apps", "users", "tokens", "blacklist", "logs", "calendar"}
	for _, view := range views {
		tmpl := template.New(view)
		// Compile layouts and views together
		files := []string{"layouts/master.html", "views/" + view + ".html"}
		parsedTmpl, err := tmpl.ParseFS(subFS, files...)
		if err != nil {
			panic("failed to compile template " + view + ": " + err.Error())
		}
		h.templates[view] = parsedTmpl
	}
}

// Render outputs cached templates executing master layouts and passing dynamic data
func (h *BaseHandler) Render(w http.ResponseWriter, view string, data interface{}) {
	tmpl, exists := h.templates[view]
	if !exists {
		utils.Errorf("Template not found: %s", view)
		http.Error(w, "Template not found: "+view, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		utils.Errorf("Failed to render layout: %s", err.Error())
		http.Error(w, "Failed to render layout: "+err.Error(), http.StatusInternalServerError)
	}
}

// HTTPError logs the HTTP error message and writes the error to the response
func (h *BaseHandler) HTTPError(w http.ResponseWriter, r *http.Request, error string, code int) {
	utils.Errorf("HTTP %d error for %s %s: %s", code, r.Method, r.URL.Path, error)
	http.Error(w, error, code)
}
