// Package checkgrp maintains the group of handlers for health checking.
package checkgrp

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/ezratameno/yify/business/sys/database"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// Handler manages the set of check endpoints.
type Handlers struct {
	Build string
	Log   *logrus.Entry
	DB    *sqlx.DB
}

// Readiness checks if the database is ready and if not will return a 500 status code.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
func (h Handlers) Readiness(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	status := "ok"
	statusCode := http.StatusOK
	if err := database.StatusCheck(ctx, h.DB); err != nil {
		status = "db not ready"
		statusCode = http.StatusInternalServerError
	}
	data := struct {
		Build  string
		Status string `json:"status"`
	}{
		Build:  h.Build,
		Status: status,
	}
	if err := response(w, statusCode, data); err != nil {
		h.Log.WithFields(logrus.Fields{
			"ERROR": err,
		}).Error("rediness")
	}
	h.Log.WithFields(logrus.Fields{
		"statusCode": statusCode,
		"method":     r.Method,
		"path":       r.URL.Path,
		"remoteaddr": r.RemoteAddr,
	}).Info("readiness")

}

// Liveness returns simple status info if the service is alive.
// If the app is deployed to a k8s cluster, it will also return pod, node,
//
//	and namespace details via the Downward API.
//
// The k8s environment variables need to be set within your POD/Deployment manifest.
func (h Handlers) Liveness(w http.ResponseWriter, r *http.Request) {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	data := struct {
		Status    string `json:"status,omit,empty"`
		Build     string `json:"build,omitempty"`
		Host      string `json:"host,omitempty"`
		Pod       string `json:"pod,omitempty"`
		PodIp     string `json:"podIP,omitempty"`
		Node      string `json:"node,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Status:    "up",
		Build:     h.Build,
		Host:      host,
		Pod:       os.Getenv("KUBERNETES_PODNAME"),
		PodIp:     os.Getenv("KUBERNETES_NAMESPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}
	statusCode := http.StatusOK
	if err := response(w, statusCode, data); err != nil {
		h.Log.WithFields(logrus.Fields{
			"ERROR": err,
		}).Error("liveness")
	}
	h.Log.WithFields(logrus.Fields{
		"statusCode": statusCode,
		"method":     r.Method,
		"path":       r.URL.Path,
		"remoteaddr": r.RemoteAddr,
	}).Info("liveness")

}

func response(w http.ResponseWriter, statusCode int, data interface{}) error {

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(statusCode)

	// Send the result back to the client.
	if _, err := w.Write(jsonData); err != nil {
		return err
	}
	return nil
}
