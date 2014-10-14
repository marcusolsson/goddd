package render

import (
	"encoding/xml"
	"html/template"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type Greeting struct {
	One string `json:"one"`
	Two string `json:"two"`
}

type GreetingXML struct {
	XMLName xml.Name `xml:"greeting"`
	One     string   `xml:"one,attr"`
	Two     string   `xml:"two,attr"`
}

func TestRenderJSON(t *testing.T) {
	render := New(Options{
	// nothing here to configure
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, 299, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), `{"one":"hello","two":"world"}`)
}

func TestRenderJSONPrefix(t *testing.T) {
	prefix := ")]}',\n"
	render := New(Options{
		PrefixJSON: []byte(prefix),
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, 300, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), prefix+`{"one":"hello","two":"world"}`)
}

func TestRenderIndentedJSON(t *testing.T) {
	render := New(Options{
		IndentJSON: true,
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, http.StatusOK, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), `{
  "one": "hello",
  "two": "world"
}`)
}

func TestRenderJSONWithError(t *testing.T) {
	render := New(Options{
	// nothing here to configure
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, 299, math.NaN())
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 500)
}

func TestRenderXML(t *testing.T) {
	render := New(Options{
	// nothing here to configure
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.XML(w, 299, GreetingXML{One: "hello", Two: "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentXML+"; charset=UTF-8")
	expect(t, res.Body.String(), `<greeting one="hello" two="world"></greeting>`)
}

func TestRenderXMLPrefix(t *testing.T) {
	prefix := "<?xml version='1.0' encoding='UTF-8'?>\n"
	render := New(Options{
		PrefixXML: []byte(prefix),
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.XML(w, 300, GreetingXML{One: "hello", Two: "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentXML+"; charset=UTF-8")
	expect(t, res.Body.String(), prefix+`<greeting one="hello" two="world"></greeting>`)
}

func TestRenderIndentedXML(t *testing.T) {
	render := New(Options{
		IndentXML: true,
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.XML(w, http.StatusOK, GreetingXML{One: "hello", Two: "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), ContentXML+"; charset=UTF-8")
	expect(t, res.Body.String(), `<greeting one="hello" two="world"></greeting>`)
}

func TestRenderXMLWithError(t *testing.T) {
	render := New(Options{
	// nothing here to configure
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.XML(w, 299, map[string]string{"foo": "bar"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 500)
}

func TestRenderBadHTML(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "nope", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 500)
	expect(t, res.Body.String(), "html/template: \"nope\" is undefined\n")
}

func TestRenderHTML(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello gophers</h1>\n")
}

func TestRenderXHTML(t *testing.T) {
	render := New(Options{
		Directory:       "fixtures/basic",
		HTMLContentType: ContentXHTML,
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentXHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello gophers</h1>\n")
}

func TestRenderExtensions(t *testing.T) {
	render := New(Options{
		Directory:  "fixtures/basic",
		Extensions: []string{".tmpl", ".html"},
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "hypertext", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "Hypertext!\n")
}

func TestRenderFuncs(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/custom_funcs",
		Funcs: []template.FuncMap{
			{
				"myCustomFunc": func() string {
					return "My custom function"
				},
			},
		},
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "index", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Body.String(), "My custom function\n")
}

func TestRenderLayout(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
		Layout:    "layout",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "content", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Body.String(), "head\n<h1>gophers</h1>\n\nfoot\n")
}

func TestRenderLayoutCurrent(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
		Layout:    "current_layout",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "content", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Body.String(), "content head\n<h1>gophers</h1>\n\ncontent foot\n")
}

func TestRenderNestedHTML(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "admin/index", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Admin gophers</h1>\n")
}

func TestRenderBadPathHTML(t *testing.T) {
	render := New(Options{
		Directory: "../../../../../../../../../../../../../../../../fixtures/basic",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 500)
}

func TestRenderDelimiters(t *testing.T) {
	render := New(Options{
		Delims:    Delims{"{[{", "}]}"},
		Directory: "fixtures/basic",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "delims", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello gophers</h1>")
}

func TestRenderBinaryData(t *testing.T) {
	render := New(Options{
	// nothing here to configure
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.Data(w, 299, []byte("hello there"))
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentBinary)
	expect(t, res.Body.String(), "hello there")
}

func TestRenderBinaryDataCustomMimeType(t *testing.T) {
	render := New(Options{
	// nothing here to configure
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContentType, "image/jpeg")
		render.Data(w, http.StatusOK, []byte("..jpeg data.."))
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), "image/jpeg")
	expect(t, res.Body.String(), "..jpeg data..")
}

func TestRenderCharsetJSON(t *testing.T) {
	render := New(Options{
		Charset: "foobar",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, 300, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=foobar")
	expect(t, res.Body.String(), `{"one":"hello","two":"world"}`)
}

func TestRenderDefaultCharsetHTML(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	// ContentLength should be deferred to the ResponseWriter and not Render
	expect(t, res.Header().Get(ContentLength), "")
	expect(t, res.Body.String(), "<h1>Hello gophers</h1>\n")
}

func TestRenderOverrideLayout(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
		Layout:    "layout",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "content", "gophers", HTMLOptions{
			Layout: "another_layout",
		})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "another head\n<h1>gophers</h1>\n\nanother foot\n")
}

func TestRenderNoRace(t *testing.T) {
	// This test used to fail if run with -race
	render := New(Options{
		Directory: "fixtures/basic",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	done := make(chan bool)
	doreq := func() {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/foo", nil)

		h.ServeHTTP(res, req)

		expect(t, res.Code, 200)
		expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
		// ContentLength should be deferred to the ResponseWriter and not Render
		expect(t, res.Header().Get(ContentLength), "")
		expect(t, res.Body.String(), "<h1>Hello gophers</h1>\n")
		done <- true
	}
	// Run two requests to check there is no race condition
	go doreq()
	go doreq()
	<-done
	<-done
}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
