package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/mgo.v2"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"

	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/inmem"
	"github.com/marcusolsson/goddd/inspection"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/mongo"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/tracking"
	"github.com/marcusolsson/goddd/voyage"
)

const (
	defaultPort              = "8080"
	defaultRoutingServiceURL = "http://localhost:7878"
	defaultMongoDBURL        = "127.0.0.1"
	defaultDBName            = "dddsample"
)

func main() {
	var (
		addr   = envString("PORT", defaultPort)
		rsurl  = envString("ROUTINGSERVICE_URL", defaultRoutingServiceURL)
		dburl  = envString("MONGODB_URL", defaultMongoDBURL)
		dbname = envString("DB_NAME", defaultDBName)

		httpAddr          = flag.String("http.addr", ":"+addr, "HTTP listen address")
		routingServiceURL = flag.String("service.routing", rsurl, "routing service URL")
		mongoDBURL        = flag.String("db.url", dburl, "MongoDB URL")
		databaseName      = flag.String("db.name", dbname, "MongoDB database name")
		inmemory          = flag.Bool("inmem", false, "use in-memory repositories")

		ctx = context.Background()
	)

	flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	// Setup repositories
	var (
		cargos         cargo.Repository
		locations      location.Repository
		voyages        voyage.Repository
		handlingEvents cargo.HandlingEventRepository
	)

	if *inmemory {
		cargos = inmem.NewCargoRepository()
		locations = inmem.NewLocationRepository()
		voyages = inmem.NewVoyageRepository()
		handlingEvents = inmem.NewHandlingEventRepository()
	} else {
		session, err := mgo.Dial(*mongoDBURL)
		if err != nil {
			panic(err)
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		cargos, _ = mongo.NewCargoRepository(*databaseName, session)
		locations, _ = mongo.NewLocationRepository(*databaseName, session)
		voyages, _ = mongo.NewVoyageRepository(*databaseName, session)
		handlingEvents = mongo.NewHandlingEventRepository(*databaseName, session)
	}

	// Configure some questionable dependencies.
	var (
		handlingEventFactory = cargo.HandlingEventFactory{
			CargoRepository:    cargos,
			VoyageRepository:   voyages,
			LocationRepository: locations,
		}
		handlingEventHandler = handling.NewEventHandler(
			inspection.NewService(cargos, handlingEvents, nil),
		)
	)

	// Facilitate testing by adding some cargos.
	storeTestData(cargos)

	fieldKeys := []string{"method"}

	var rs routing.Service
	rs = routing.NewProxyingMiddleware(ctx, *routingServiceURL)(rs)

	var bs booking.Service
	bs = booking.NewService(cargos, locations, handlingEvents, rs)
	bs = booking.NewLoggingService(log.With(logger, "component", "booking"), bs)
	bs = booking.NewInstrumentingService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "booking_service",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "booking_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		bs,
	)

	var ts tracking.Service
	ts = tracking.NewService(cargos, handlingEvents)
	ts = tracking.NewLoggingService(log.With(logger, "component", "tracking"), ts)
	ts = tracking.NewInstrumentingService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "tracking_service",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "tracking_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		ts,
	)

	var hs handling.Service
	hs = handling.NewService(handlingEvents, handlingEventFactory, handlingEventHandler)
	hs = handling.NewLoggingService(log.With(logger, "component", "handling"), hs)
	hs = handling.NewInstrumentingService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "handling_service",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "handling_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		hs,
	)

	httpLogger := log.With(logger, "component", "http")

	mux := http.NewServeMux()

	mux.Handle("/booking/v1/", booking.MakeHandler(bs, httpLogger))
	mux.Handle("/tracking/v1/", tracking.MakeHandler(ts, httpLogger))
	mux.Handle("/handling/v1/", handling.MakeHandler(hs, httpLogger))

	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())

	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
		errs <- http.ListenAndServe(*httpAddr, nil)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

func storeTestData(r cargo.Repository) {
	test1 := cargo.New("FTL456", cargo.RouteSpecification{
		Origin:          location.AUMEL,
		Destination:     location.SESTO,
		ArrivalDeadline: time.Now().AddDate(0, 0, 7),
	})
	if err := r.Store(test1); err != nil {
		panic(err)
	}

	test2 := cargo.New("ABC123", cargo.RouteSpecification{
		Origin:          location.SESTO,
		Destination:     location.CNHKG,
		ArrivalDeadline: time.Now().AddDate(0, 0, 14),
	})
	if err := r.Store(test2); err != nil {
		panic(err)
	}
}
