package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/oceans2001/warden/internal/hubble"
	"github.com/oceans2001/warden/internal/k8s"
	"github.com/oceans2001/warden/internal/metrics"
	"github.com/oceans2001/warden/internal/policy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CreateAgentRequest struct {
	Name           string   `json:"name"`
	Image          string   `json:"image"`
	Command        []string `json:"command"`
	AllowedDomains []string `json:"allowed_domains"`
}

func refreshMetricsFromStore(store *hubble.Store, sessions []string) {
	for _, name := range sessions {
		events := store.Get(name)

		counts := make(map[string]int)
		for _, e := range events {
			key := e.Verdict + "|" + e.Destination
			counts[key]++
		}

		for key, count := range counts {
			verdict, destination := splitKey(key)
			metrics.FlowGauge.WithLabelValues(name, verdict, destination).Set(float64(count))
		}
	}
}

func splitKey(key string) (string, string) {
	for i := 0; i < len(key); i++ {
		if key[i] == '|' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}

func main() {
	clientset, err := k8s.NewClient()
	if err != nil {
		fmt.Println("Error connecting to cluster:", err)
		return
	}

	dynamicClient, err := k8s.NewDynamicClient()
	if err != nil {
		fmt.Println("Error creating dynamic client:", err)
		return
	}

	store := hubble.NewStore(100)
	go hubble.StartStream(store)

	http.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		var req CreateAgentRequest
		json.NewDecoder(r.Body).Decode(&req)

		command := req.Command
		if len(command) == 0 {
			command = []string{"sh", "-c", "sleep 3600"}
		}

		err := k8s.CreatePod(clientset, req.Name, req.Image, command)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		pol := policy.BuildFQDNPolicy(req.Name, req.AllowedDomains)
		err = policy.ApplyPolicy(dynamicClient, pol)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "Agent %s created\n", req.Name)
	})

	http.HandleFunc("/agents/list", func(w http.ResponseWriter, r *http.Request) {
		names, err := k8s.ListAgents(clientset)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		json.NewEncoder(w).Encode(names)
	})

	http.HandleFunc("/agents/delete", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "name query parameter required", 400)
			return
		}

		err := k8s.DeletePod(clientset, name)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = policy.DeletePolicy(dynamicClient, name)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "Agent %s deleted\n", name)
	})

	http.HandleFunc("/agents/flows", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "name query parameter required", 400)
			return
		}

		events := store.Get(name)
		json.NewEncoder(w).Encode(events)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		names, err := k8s.ListAgents(clientset)
		if err == nil {
			refreshMetricsFromStore(store, names)
		}
		promhttp.Handler().ServeHTTP(w, r)
	})

	fmt.Println("Warden API starting on :8080")
	http.ListenAndServe(":8080", nil)
}
