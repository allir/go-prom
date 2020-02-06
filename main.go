package main

import (
	"fmt"
	"log"
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

	bingoChain = promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(requestDuration.MustCurryWith(prometheus.Labels{"handler": "bingo"}),
		promhttp.InstrumentHandlerCounter(counter,
			promhttp.InstrumentHandlerResponseSize(responseSize, Bingo()),
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
	router.Handle("/bingo", bingoChain)
	router.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", router))
}

// SayHello says hi
func SayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusBadRequest // if req is not GET
		if r.Method == "GET" {
			code = http.StatusOK
			vars := mux.Vars(r)
			name := vars["name"]

			if name == "" {
			    log.Println(r.URL.Path + " - GET - 200 - Saying hello")
			} else {
				log.Println(r.URL.Path + " - GET - 200 - Saying hello to " + name)
			}

			greet := fmt.Sprintf("Hello %s \n", name)
			
			w.Write([]byte(greet))
		} else {
			w.WriteHeader(code)
		}
	}
}

// Bingo says BINGO!
func Bingo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusBadRequest // if req is not GET
		if r.Method == "GET" {
			code = http.StatusOK

			log.Println(r.URL.Path + " - GET - 200 - BINGO")

			greet := fmt.Sprintf("BINGO!\n")
			
			w.Write([]byte(greet))
		} else {
			w.WriteHeader(code)
		}
	}
}
