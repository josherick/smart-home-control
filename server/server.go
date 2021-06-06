package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Logger interface {
	Info(message string)
	Critical(message string)
	CriticalError(err error)
	WrapHandlerWithAccessLog(mux http.Handler) http.Handler
}

type InfoLogFilenameProvider interface {
	Filename(t time.Time) string
}

type PlugController interface {
	TurnOnCorrespondingPlug(sensorID string) error
	TurnOffCorrespondingPlug(sensorID string) error
}

type server struct {
	port int

	logger Logger
	pc     PlugController
	fp     InfoLogFilenameProvider
}

func New(
	port int,
	logger Logger,
	pc PlugController,
	fp InfoLogFilenameProvider,
) *server {
	return &server{port, logger, pc, fp}
}

func (s *server) Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/outside_bounds", s.handleTempOutsideBounds)
	mux.HandleFunc("/logs", s.handleTempOutsideBounds)
	muxWithAccessLogging := s.logger.WrapHandlerWithAccessLog(mux)

	port := fmt.Sprintf(":%d", s.port)
	s.logger.CriticalError(http.ListenAndServe(port, muxWithAccessLogging))
}

func (s *server) handleTempOutsideBounds(w http.ResponseWriter, r *http.Request) {
	// Validate query params
	qs := r.URL.Query()
	sensorID := qs.Get("sensor_id")
	turnOnStr := qs.Get("turn_on")
	tempC := qs.Get("temp_c")
	if sensorID == "" || (turnOnStr != "false" && turnOnStr != "true") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	turnOn := turnOnStr == "true"
	tempCFloat, _ := strconv.ParseFloat(tempC, 64)
	tempFFloat := tempCFloat*(9.0/5.0) + 32

	// Log event
	var onOffStr = "off"
	if turnOn {
		onOffStr = "on"
	}
	msg := fmt.Sprintf(
		"[%s] Request to turn %s corresponding plug. Temp: %.2f C / %.2f F",
		sensorID,
		onOffStr,
		tempCFloat,
		tempFFloat,
	)
	s.logger.Info(msg)

	var err error
	if turnOn {
		err = s.pc.TurnOnCorrespondingPlug(sensorID)
	} else {
		err = s.pc.TurnOffCorrespondingPlug(sensorID)
	}
	if err != nil {
		s.logger.CriticalError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	msg = fmt.Sprintf(
		"[%s] Successfully fulfilled request to turn %s corresponding plug.",
		sensorID,
		onOffStr,
	)
	s.logger.Info(msg)
}

func (s *server) getLog(w http.ResponseWriter, r *http.Request) {
	ds := r.URL.Query().Get("ds")

	// Use a valid date string or today.
	var t time.Time
	if ds != "" {
		parsed, err := time.Parse(ds, "2006-01-02")
		if err != nil {
			t = parsed
		}
	}
	if (t == time.Time{}) {
		t = time.Now()
	}

	filename := s.fp.Filename(t)
	body, err := os.ReadFile(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(body)
}
