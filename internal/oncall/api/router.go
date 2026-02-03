package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vjranagit/grafana/internal/oncall/store"
)

func NewRouter(st *store.Store) chi.Router {
	r := chi.NewRouter()

	h := &handlers{
		store:          st,
		alertProcessor: NewAlertProcessor(st),
	}

	// Schedules
	r.Route("/schedules", func(r chi.Router) {
		r.Get("/", h.listSchedules)
		r.Post("/", h.createSchedule)
		r.Get("/{id}", h.getSchedule)
		r.Put("/{id}", h.updateSchedule)
		r.Delete("/{id}", h.deleteSchedule)
		r.Get("/{id}/oncall", h.getCurrentOnCall)
	})

	// Escalation Chains
	r.Route("/escalations", func(r chi.Router) {
		r.Get("/", h.listEscalationChains)
		r.Post("/", h.createEscalationChain)
		r.Get("/{id}", h.getEscalationChain)
		r.Put("/{id}", h.updateEscalationChain)
		r.Delete("/{id}", h.deleteEscalationChain)
	})

	// Alerts (webhook receivers)
	r.Route("/alerts", func(r chi.Router) {
		r.Post("/prometheus", h.receivePrometheusAlert)
		r.Post("/grafana", h.receiveGrafanaAlert)
		r.Post("/webhook", h.receiveWebhookAlert)
		r.Get("/", h.listAlerts)
		r.Get("/{id}", h.getAlert)
		r.Post("/{id}/acknowledge", h.acknowledgeAlert)
		r.Post("/{id}/resolve", h.resolveAlert)
	})

	// Integrations
	r.Route("/integrations", func(r chi.Router) {
		r.Get("/", h.listIntegrations)
		r.Post("/", h.createIntegration)
		r.Get("/{id}", h.getIntegration)
		r.Delete("/{id}", h.deleteIntegration)
	})

	return r
}

type handlers struct {
	store          *store.Store
	alertProcessor *AlertProcessor
}

// Placeholder handlers - to be implemented
func (h *handlers) listSchedules(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, []interface{}{})
}

func (h *handlers) createSchedule(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *handlers) getSchedule(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"id": chi.URLParam(r, "id")})
}

func (h *handlers) updateSchedule(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *handlers) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) getCurrentOnCall(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"schedule_id": chi.URLParam(r, "id"),
		"oncall_user": "user123",
	})
}

func (h *handlers) listEscalationChains(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, []interface{}{})
}

func (h *handlers) createEscalationChain(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *handlers) getEscalationChain(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"id": chi.URLParam(r, "id")})
}

func (h *handlers) updateEscalationChain(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *handlers) deleteEscalationChain(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Real implementation for Prometheus alerts
func (h *handlers) receivePrometheusAlert(w http.ResponseWriter, r *http.Request) {
	var webhook PrometheusWebhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		slog.Error("failed to decode prometheus webhook", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	slog.Info("received prometheus webhook",
		"status", webhook.Status,
		"alerts", len(webhook.Alerts))

	alertGroups, err := h.alertProcessor.ProcessPrometheusWebhook(&webhook)
	if err != nil {
		slog.Error("failed to process alerts", "error", err)
		http.Error(w, "failed to process alerts", http.StatusInternalServerError)
		return
	}

	slog.Info("processed alerts",
		"count", len(alertGroups),
		"status", webhook.Status)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "received",
		"alerts_count":  len(alertGroups),
		"webhook_status": webhook.Status,
	})
}

func (h *handlers) receiveGrafanaAlert(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse Grafana alert webhook format
	respondJSON(w, http.StatusOK, map[string]string{"status": "received"})
}

func (h *handlers) receiveWebhookAlert(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse generic webhook format
	respondJSON(w, http.StatusOK, map[string]string{"status": "received"})
}

func (h *handlers) listAlerts(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, []interface{}{})
}

func (h *handlers) getAlert(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"id": chi.URLParam(r, "id")})
}

func (h *handlers) acknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

func (h *handlers) resolveAlert(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (h *handlers) listIntegrations(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, []interface{}{})
}

func (h *handlers) createIntegration(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *handlers) getIntegration(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"id": chi.URLParam(r, "id")})
}

func (h *handlers) deleteIntegration(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
