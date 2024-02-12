package core

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
)

// Definition of Exporter struct for expose scrapped metrics
type Exporter struct {
	Ctx      context.Context
	host     string
	path     string
	port     int
	interval int
	requests *prometheus.GaugeVec
	Server   *http.Server
	Mux      *http.ServeMux
}

// Initialize a exporter strucs
func NewExporter(ctx context.Context, host string, port int, path string, interval int) *Exporter {
	uri := fmt.Sprintf("%s:%d", host, port)
	mux := http.NewServeMux()
	exp := &Exporter{
		host:     host,
		port:     port,
		path:     path,
		interval: interval,
		Server:   &http.Server{Addr: uri, Handler: mux},
		Mux:      mux,
		// Todo : handle this logic directly in the service
		requests: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fizzbuzz_request_sec",
				Help: "Number of requests",
			}, []string{"code", "method", "path"}),
	}

	return exp
}

// Not used yet
// Watch metrics in the interval time
func (o *Exporter) WatchedMetrics() {
	go func() {
		for {
			// What you whant to watch
			time.Sleep(time.Duration(o.interval) * time.Second)
		}
	}()
}

// Start to expose /metrics
func (o *Exporter) Start() {
	o.Mux.Handle(o.path, promhttp.Handler())
	go func() {
		log.Printf("[EXPORTER] Start listen on %s:%d\n", o.host, o.port)
		err := o.Server.ListenAndServe()
		if err != nil {
			log.Println("[EXPORTER] Error listen: ", err)
		}
	}()
}

// Increment Gauge request by code, method and path
func (o *Exporter) IncrRequests(code int, method string, path string) {
	str_code := strconv.Itoa(code)
	o.requests.WithLabelValues(str_code, method, path).Inc()
}

// Http Gateway middleware for handle metrics on all requests services
func (o *Exporter) HandleHttpHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		path := r.URL.Path
		rwh := NewResponseWriterHandler(w)
		h.ServeHTTP(rwh, r)
		o.IncrRequests(rwh.StatusCode, method, path)
	})
}

func (o *Exporter) GracefulStop() error {
	log.Println("[EXPORTER] Graceful Stop")
	return o.Server.Shutdown(o.Ctx)
}
