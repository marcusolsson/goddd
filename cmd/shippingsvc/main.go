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

	"github.com/globalsign/mgo"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/inmem"
	"github.com/marcusolsson/goddd/inspection"
	"github.com/marcusolsson/goddd/mongo"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/server"
	"github.com/marcusolsson/goddd/tracking"
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
		cargos         shipping.CargoRepository
		locations      shipping.LocationRepository
		voyages        shipping.VoyageRepository
		handlingEvents shipping.HandlingEventRepository
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
		handlingEventFactory = shipping.HandlingEventFactory{
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

	var rs shipping.RoutingService
	rs = routing.NewProxyMiddleware(ctx, *routingServiceURL)(rs)

	var bs booking.Service
	bs = booking.NewService(cargos, locations, handlingEvents, rs)
	bs = booking.NewLoggingMiddleware(log.With(logger, "component", "booking"), bs)
	bs = booking.NewInstrumentingMiddleware(
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
	ts = tracking.NewLoggingMiddleware(log.With(logger, "component", "tracking"), ts)
	ts = tracking.NewInstrumentingMiddleware(
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
	hs = handling.NewLoggingMiddleware(log.With(logger, "component", "handling"), hs)
	hs = handling.NewInstrumentingMiddleware(
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

	srv := server.New(bs, ts, hs, log.With(logger, "component", "http"))

	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
		errs <- http.ListenAndServe(*httpAddr, srv)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

func storeTestData(r shipping.CargoRepository) {
	test1 := shipping.NewCargo("FTL456", shipping.RouteSpecification{
		Origin:          shipping.AUMEL,
		Destination:     shipping.SESTO,
		ArrivalDeadline: time.Now().AddDate(0, 0, 7),
	})
	if err := r.Store(test1); err != nil {
		panic(err)
	}

	test2 := shipping.NewCargo("ABC123", shipping.RouteSpecification{
		Origin:          shipping.SESTO,
		Destination:     shipping.CNHKG,
		ArrivalDeadline: time.Now().AddDate(0, 0, 14),
	})
	if err := r.Store(test2); err != nil {
		panic(err)
	}
}
