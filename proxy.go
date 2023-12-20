package spanner

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

const (
	defaultNodeExporterPort = 9100
)

type Target struct {
	Labels  map[string]string `json:"labels"`
	Targets []string          `json:"targets"`
}

func jobListToTargets(jobList []Job) []Target {
	targets := []Target{}
	for _, job := range jobList {
		if len(job.Nodes) == 0 {
			continue
		}
		target := Target{
			Labels: map[string]string{
				"job": job.Name,
			},
			Targets: []string{},
		}
		for _, node := range job.Nodes {
			node = RewriteNode(node)
			nodeURL := fmt.Sprintf("%s:%d", node, defaultNodeExporterPort)
			target.Targets = append(target.Targets, nodeURL)
		}
		targets = append(targets, target)
	}
	return targets
}

func PrometheusTargetsHandler(bs BatchSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobList, err := bs.ListJobs(false)
		if err != nil {
			log.Err(err).Msg("query job list")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		targets := jobListToTargets(jobList)
		if err := json.NewEncoder(w).Encode(targets); err != nil {
			log.Err(err).Msg("encode targets")
			return
		}
	}
}

func Proxy(bs BatchSystem, addr string) error {
	proxy := goproxy.NewProxyHttpServer()

	r := mux.NewRouter()
	r.HandleFunc("/targets", PrometheusTargetsHandler(bs))
	r.PathPrefix("/").Handler(proxy)

	log.Info().Msgf("starting proxy server at %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal().Err(err).Msg("listen and serve")
	}
	return nil
}
