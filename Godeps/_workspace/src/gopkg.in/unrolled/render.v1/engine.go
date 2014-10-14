package render

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"net/http"
)

// Engine is the generic interface for all responses.
type Engine interface {
	Render(http.ResponseWriter, interface{}) error
}

// Head defines the basic ContentType and Status fields.
type Head struct {
	ContentType string
	Status      int
}

// XML built-in renderer.
type XML struct {
	Head
	Indent bool
	Prefix []byte
}

// JSON built-in renderer.
type JSON struct {
	Head
	Indent bool
	Prefix []byte
}

// HTML built-in renderer.
type HTML struct {
	Head
	Name      string
	Templates *template.Template
}

// Data built-in renderer.
type Data struct {
	Head
}

// Write outputs the header content.
func (h Head) Write(w http.ResponseWriter) {
	w.Header().Set(ContentType, h.ContentType)
	w.WriteHeader(h.Status)
}

// Render a data response.
func (d Data) Render(w http.ResponseWriter, v interface{}) error {
	c := w.Header().Get(ContentType)
	if c != "" {
		d.Head.ContentType = c
	}

	d.Head.Write(w)
	w.Write(v.([]byte))
	return nil
}

// Render a JSON response.
func (j JSON) Render(w http.ResponseWriter, v interface{}) error {
	var result []byte
	var err error

	if j.Indent {
		result, err = json.MarshalIndent(v, "", "  ")
	} else {
		result, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}

	// JSON marshaled fine, write out the result.
	j.Head.Write(w)
	if len(j.Prefix) > 0 {
		w.Write(j.Prefix)
	}
	w.Write(result)
	return nil
}

// Render an XML response.
func (x XML) Render(w http.ResponseWriter, v interface{}) error {
	var result []byte
	var err error

	if x.Indent {
		result, err = xml.MarshalIndent(v, "", "  ")
	} else {
		result, err = xml.Marshal(v)
	}
	if err != nil {
		return err
	}

	// XML marshaled fine, write out the result.
	x.Head.Write(w)
	if len(x.Prefix) > 0 {
		w.Write(x.Prefix)
	}
	w.Write(result)
	return nil
}

// Render a HTML response.
func (h HTML) Render(w http.ResponseWriter, binding interface{}) error {
	out := new(bytes.Buffer)
	err := h.Templates.ExecuteTemplate(out, h.Name, binding)
	if err != nil {
		return err
	}

	h.Head.Write(w)
	w.Write(out.Bytes())
	return nil
}
