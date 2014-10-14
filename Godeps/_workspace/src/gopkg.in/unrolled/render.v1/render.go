package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ContentType header constant.
	ContentType = "Content-Type"
	// ContentLength header constant.
	ContentLength = "Content-Length"
	// ContentBinary header value for binary data.
	ContentBinary = "application/octet-stream"
	// ContentJSON header value for JSON data.
	ContentJSON = "application/json"
	// ContentHTML header value for HTML data.
	ContentHTML = "text/html"
	// ContentXHTML header value for XHTML data.
	ContentXHTML = "application/xhtml+xml"
	// ContentXML header value for XML data.
	ContentXML = "text/xml"
	// Default character encoding.
	defaultCharset = "UTF-8"
)

// Included helper functions for use when rendering html.
var helperFuncs = template.FuncMap{
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	"current": func() (string, error) {
		return "", nil
	},
}

// Delims represents a set of Left and Right delimiters for HTML template rendering.
type Delims struct {
	// Left delimiter, defaults to {{.
	Left string
	// Right delimiter, defaults to }}.
	Right string
}

// Options is a struct for specifying configuration options for the render.Render object.
type Options struct {
	// Directory to load templates. Default is "templates".
	Directory string
	// Layout template name. Will not render a layout if blank (""). Defaults to blank ("").
	Layout string
	// Extensions to parse template files from. Defaults to [".tmpl"].
	Extensions []string
	// Funcs is a slice of FuncMaps to apply to the template upon compilation. This is useful for helper functions. Defaults to [].
	Funcs []template.FuncMap
	// Delims sets the action delimiters to the specified strings in the Delims struct.
	Delims Delims
	// Appends the given character set to the Content-Type header. Default is "UTF-8".
	Charset string
	// Outputs human readable JSON.
	IndentJSON bool
	// Outputs human readable XML.
	IndentXML bool
	// Prefixes the JSON output with the given bytes.
	PrefixJSON []byte
	// Prefixes the XML output with the given bytes.
	PrefixXML []byte
	// Allows changing of output to XHTML instead of HTML. Default is "text/html"
	HTMLContentType string
	// If IsDevelopment is set to true, this will recompile the templates on every request. Default if false.
	IsDevelopment bool
}

// HTMLOptions is a struct for overriding some rendering Options for specific HTML call.
type HTMLOptions struct {
	// Layout template name. Overrides Options.Layout.
	Layout string
}

// Render is a service that provides functions for easily writing JSON, XML,
// Binary Data, and HTML templates out to a http Response.
type Render struct {
	// Customize Secure with an Options struct.
	opt             Options
	templates       *template.Template
	compiledCharset string
}

// New constructs a new Render instance with the supplied options.
func New(options Options) *Render {
	r := Render{
		opt: options,
	}

	r.prepareOptions()
	r.compileTemplates()

	return &r
}

func (r *Render) prepareOptions() {
	// Fill in the defaults if need be.
	if len(r.opt.Charset) == 0 {
		r.opt.Charset = defaultCharset
	}
	r.compiledCharset = "; charset=" + r.opt.Charset

	if len(r.opt.Directory) == 0 {
		r.opt.Directory = "templates"
	}
	if len(r.opt.Extensions) == 0 {
		r.opt.Extensions = []string{".tmpl"}
	}
	if len(r.opt.HTMLContentType) == 0 {
		r.opt.HTMLContentType = ContentHTML
	}
}

func (r *Render) compileTemplates() {
	dir := r.opt.Directory
	r.templates = template.New(dir)
	r.templates.Delims(r.opt.Delims.Left, r.opt.Delims.Right)

	// Walk the supplied directory and compile any files that match our extension list.
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = "." + strings.Join(strings.Split(rel, ".")[1:], ".")
		}

		for _, extension := range r.opt.Extensions {
			if ext == extension {

				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				name := (rel[0 : len(rel)-len(ext)])
				tmpl := r.templates.New(filepath.ToSlash(name))

				// Add our funcmaps.
				for _, funcs := range r.opt.Funcs {
					tmpl.Funcs(funcs)
				}

				// Break out if this parsing fails. We don't want any silent server starts.
				template.Must(tmpl.Funcs(helperFuncs).Parse(string(buf)))
				break
			}
		}

		return nil
	})
}

// Render is the generic function called by XML, JSON, Data, HTML, and can be called by custom implementations.
func (r *Render) Render(w http.ResponseWriter, e Engine, data interface{}) {
	err := e.Render(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// XML marshals the given interface object and writes the XML response.
func (r *Render) XML(w http.ResponseWriter, status int, v interface{}) {
	head := Head{
		ContentType: ContentXML + r.compiledCharset,
		Status:      status,
	}

	x := XML{
		Head:   head,
		Indent: r.opt.IndentXML,
		Prefix: r.opt.PrefixXML,
	}

	r.Render(w, x, v)
}

// JSON marshals the given interface object and writes the JSON response.
func (r *Render) JSON(w http.ResponseWriter, status int, v interface{}) {
	head := Head{
		ContentType: ContentJSON + r.compiledCharset,
		Status:      status,
	}

	j := JSON{
		Head:   head,
		Indent: r.opt.IndentJSON,
		Prefix: r.opt.PrefixJSON,
	}

	r.Render(w, j, v)
}

// Data writes out the raw bytes as binary data.
func (r *Render) Data(w http.ResponseWriter, status int, v []byte) {
	head := Head{
		ContentType: ContentBinary,
		Status:      status,
	}

	d := Data{
		Head: head,
	}

	r.Render(w, d, v)
}

// HTML builds up the response from the specified template and bindings.
func (r *Render) HTML(w http.ResponseWriter, status int, name string, binding interface{}, htmlOpt ...HTMLOptions) {
	// If we are in development mode, recompile the templates on every HTML request.
	if r.opt.IsDevelopment {
		r.compileTemplates()
	}

	opt := r.prepareHTMLOptions(htmlOpt)

	// Assign a layout if there is one.
	if len(opt.Layout) > 0 {
		r.addYield(name, binding)
		name = opt.Layout
	}

	head := Head{
		ContentType: r.opt.HTMLContentType + r.compiledCharset,
		Status:      status,
	}

	h := HTML{
		Head:      head,
		Name:      name,
		Templates: r.templates,
	}

	r.Render(w, h, binding)
}

func (r *Render) execute(name string, binding interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	return buf, r.templates.ExecuteTemplate(buf, name, binding)
}

func (r *Render) addYield(name string, binding interface{}) {
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf, err := r.execute(name, binding)
			// Return safe HTML here since we are rendering our own template.
			return template.HTML(buf.String()), err
		},
		"current": func() (string, error) {
			return name, nil
		},
	}
	r.templates.Funcs(funcs)
}

func (r *Render) prepareHTMLOptions(htmlOpt []HTMLOptions) HTMLOptions {
	if len(htmlOpt) > 0 {
		return htmlOpt[0]
	}

	return HTMLOptions{
		Layout: r.opt.Layout,
	}
}
