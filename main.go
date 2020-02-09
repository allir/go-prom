package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)


var (
	inFlightGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "hello_requests_in_flight",
			Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	counter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hello_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"handler", "code", "method"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
        	Name:    "hello_request_duration_seconds",
        	Help:    "Time (in seconds) spent serving HTTP requests.",
			//Buckets: prometheus.DefBuckets,
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "code", "method"},
	)

	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hello_response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			//Buckets: prometheus.DefBuckets,
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{"handler", "code", "method"},
	)

	// Middleware replaces this chain
	//helloChain = promhttp.InstrumentHandlerInFlight(inFlightGauge,
	//	promhttp.InstrumentHandlerDuration(requestDuration,
	//		promhttp.InstrumentHandlerCounter(counter,
	//			promhttp.InstrumentHandlerResponseSize(responseSize, SayHello()),
	//		),
	//	),
	//)
)

//func init() {
//	// Metrics have to be registered to be exposed:
//	prometheus.MustRegister(
//		inFlightGauge,
//		counter,
//		requestDuration,
//		responseSize,
//	)
//}

func MonitoringMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Println("I'm in your middleware")

		inFlightGauge.Inc()
		defer func() {
			handler := r.URL.Path
			code := strconv.Itoa(http.StatusOK)
			inFlightGauge.Dec()
			// This counter may be unneeded as the request duration histogram also has a counter
			counter.WithLabelValues(handler, code, r.Method).Inc()

			duration := time.Since(start)
			requestDuration.WithLabelValues(handler, code, r.Method).Observe(duration.Seconds())
			log.Printf("The requuest took %f seconds tos serve", duration.Seconds())
		}()

		// How to get the byte size of the responese
		//responseSize.WithLabelValues(strconv.Itoa(http.StatusOK), r.Method).Observe(float64(sizeBytes))

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("Starting server, listening on port 8080")
	router := mux.NewRouter()

	router.Handle("/hello", SayHello())
	router.Handle("/hello/{name}", SayHello())
	router.Handle("/metrics", promhttp.Handler())

	router.Use(MonitoringMiddleware)

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
