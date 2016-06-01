package tracking

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/repository"
)

func TestTrackTrackCargo(t *testing.T) {
	var (
		cargos         = repository.NewInMemcargo()
		handlingEvents = repository.NewInMemHandlingEvent()
		service        = NewService(cargos, handlingEvents)
	)

	c := cargo.New("TEST", cargo.RouteSpecification{
		Origin:          "SESTO",
		Destination:     "FIHEL",
		ArrivalDeadline: time.Date(2005, 12, 4, 0, 0, 0, 0, time.UTC),
	})
	cargos.Store(c)

	ctx := context.Background()

	logger := log.NewLogfmtLogger(ioutil.Discard)

	h := MakeHandler(ctx, service, logger)

	req, _ := http.NewRequest("GET", "http://example.com/tracking/v1/cargos/TEST", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("rec.Code = %d; want = %d", rec.Code, http.StatusOK)
	}

	if content := rec.Header().Get("Content-Type"); content != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q; want = %q", content, "application/json; charset=utf-8")
	}

	var response trackCargoResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Error(err)
	}

	if response.Err != nil {
		t.Errorf("response.Err = %q", response.Err)
	}

	var eta time.Time

	want := Cargo{
		TrackingID:           "TEST",
		Origin:               "SESTO",
		Destination:          "FIHEL",
		ArrivalDeadline:      time.Date(2005, 12, 4, 0, 0, 0, 0, time.UTC),
		ETA:                  eta.In(time.UTC),
		StatusText:           "Not received",
		NextExpectedActivity: "There are currently no expected activities for this cargo.",
		Events:               nil,
	}

	if !reflect.DeepEqual(want, *response.Cargo) {
		t.Errorf("response.Cargo = %#v; want = %#v", response.Cargo, want)
	}
}

func TestTrackUnknownCargo(t *testing.T) {
	var (
		cargos         = repository.NewInMemcargo()
		handlingEvents = repository.NewInMemHandlingEvent()
		service        = NewService(cargos, handlingEvents)
	)

	ctx := context.Background()

	logger := log.NewLogfmtLogger(ioutil.Discard)

	h := MakeHandler(ctx, service, logger)

	req, _ := http.NewRequest("GET", "http://example.com/tracking/v1/cargos/not_found", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("rec.Code = %d; want = %d", rec.Code, http.StatusNotFound)
	}

	wantContent := "application/json; charset=utf-8"
	if got := rec.Header().Get("Content-Type"); got != wantContent {
		t.Errorf("Content-Type = %q; want = %q", got, wantContent)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Error(err)
	}

	err, ok := response["error"]
	if !ok {
		t.Error("missing error")
	}
	if err != "unknown cargo" {
		t.Errorf(`"error": %q; want = %q`, err, "unknown cargo")
	}
}
