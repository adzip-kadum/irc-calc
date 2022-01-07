package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"go.uber.org/atomic"

	"github.com/adzip-kadum/irc-calc/worker"
)

//var ErrHealthCheck = errors.New("healthcheck failed")

func init() {
	RegisterLivenessProbe("liveness-probe-enabled", func() error {
		if !livenessEnabled.Load() {
			return errors.New("liveness probe disabled")
		}
		return nil
	})
	RegisterReadinessProbe("readiness-probe-enabled", func() error {
		if !readinessEnabled.Load() {
			return errors.New("readiness probe disabled")
		}
		return nil
	})
}

func EnableLivenessProbe() {
	livenessEnabled.Store(true)
}

func DisableLivenessProbe() {
	livenessEnabled.Store(false)
}

func EnableReadinessProbe() {
	readinessEnabled.Store(true)
}

func DisableReadinessProbe() {
	readinessEnabled.Store(false)
}

type Probe func() error

var (
	livenessEnabled  *atomic.Bool = atomic.NewBool(false)
	readinessEnabled *atomic.Bool = atomic.NewBool(false)

	liveness  = map[string]Probe{}
	readiness = map[string]Probe{}
)

func RegisterLivenessProbe(name string, probe Probe) {
	if _, ok := liveness[name]; ok {
		panic(fmt.Sprintf("liveness probe %s already registered", name))
	}
	liveness[name] = probe
}

func UnregisterLivenessProbe(name string) {
	delete(liveness, name)
}

func RegisterReadinessProbe(name string, probe Probe) {
	if _, ok := readiness[name]; ok {
		panic(fmt.Sprintf("readiness probe %s already registered", name))
	}
	readiness[name] = probe
}

func UnregisterReadinessProbe(name string) {
	delete(readiness, name)
}

func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	response(w, probe(liveness))
}

func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	response(w, probe(readiness))
}

type HealthCheckService struct {
	lastError error
	closer    *worker.Closer
	interval  time.Duration
	sync.RWMutex
}

func NewHealthCheckService(interval time.Duration) *HealthCheckService {
	h := &HealthCheckService{
		closer:   worker.NewCloser(context.Background(), 1),
		interval: interval,
	}
	go worker.Worker(h.closer.Context, "healthcheck", interval, h.update, nil, h.closer.WaitGroup)

	return h
}

func (h *HealthCheckService) HealthCheck() error {
	h.RLock()
	defer h.RUnlock()
	return h.lastError

}

func (h *HealthCheckService) Close() {
	h.closer.Close()
	h.Lock()
	defer h.Unlock()
	h.lastError = errors.New("healthcheck service closed")
}

func (h *HealthCheckService) update() {
	lastError := probe(liveness)
	err := probe(readiness)
	if err != nil {
		lastError = multierror.Append(lastError, err)
	}
	h.Lock()
	h.lastError = lastError
	h.Unlock()
}

func response(w http.ResponseWriter, err error) {
	if err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`"ok"`))
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(err.Error())
	}
}

func probe(probes map[string]Probe) error {
	var result error
	for name, probe := range probes {
		err := probe()
		if err != nil {
			result = multierror.Append(result, errors.Wrap(err, name))
		}
	}
	return result
}
