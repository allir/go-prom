package main

import (
	"fmt"
	"log"
	"time"
	"strconv"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/prometheus/client_golang/prometheus"
    //"github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)


var (
	inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hello_requests_in_flight",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hello_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
        	Name:    "hello_request_duration_seconds",
        	Help:    "Time (in seconds) spent serving HTTP requests.",
			//Buckets: prometheus.DefBuckets,
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler"},
	)

	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hello_response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			//Buckets: prometheus.DefBuckets,
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{},
	)

	helloChain = promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(requestDuration.MustCurryWith(prometheus.Labels{"handler": "hello"}),
		promhttp.InstrumentHandlerCounter(counter,
			promhttp.InstrumentHandlerResponseSize(responseSize, SayHello()),
			),
		),
	)

	timeChain = promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(requestDuration.MustCurryWith(prometheus.Labels{"handler": "time"}),
		promhttp.InstrumentHandlerCounter(counter,
			promhttp.InstrumentHandlerResponseSize(responseSize, Time()),
			),
		),
	)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(
		inFlightGauge,
		counter,
		requestDuration,
		responseSize,
	)
}

func main() {
	log.Println("Starting server, listening on port 8080")
	router := mux.NewRouter()
	router.Handle("/hello", helloChain)
	router.Handle("/hello/{name}", helloChain)
	router.Handle("/time/{seconds}", timeChain)
	router.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", router))
}

func SayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusBadRequest // if req is not GET
		if r.Method == "GET" {
			code = http.StatusOK
			vars := mux.Vars(r)
			name := vars["name"]

			if name == "" {
			    log.Println(r.URL.Path + " - 200 - Saying hello")
			} else {
				log.Println(r.URL.Path + " - 200 - Saying hello to " + name)
			}

			greet := fmt.Sprintf("Hello %s \n", name)
			
			w.Write([]byte(greet))
		} else {
			w.WriteHeader(code)
		}
	}
}

func Time() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusBadRequest // if req is not GET
		if r.Method == "GET" {
			code = http.StatusOK
			vars := mux.Vars(r)
			seconds, _ := strconv.Atoi(vars["seconds"])

		    log.Println(r.URL.Path + " - 200 - Time " + string(seconds))
			time.Sleep(time.Duration(seconds) * time.Second)
			greet := fmt.Sprintf("Wasted %v seconds.\n", seconds)

			w.Write([]byte(greet))
		} else {
			w.WriteHeader(code)
		}
	}
}

