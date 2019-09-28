package caddydb

import (
	"github.com/mholt/caddy"
)

func init() {
	err := connect()

	// Only register event hooks if database connection succeeds
	if err == nil {
		caddy.RegisterEventHook("caddydb-cert-failure", onDemandCertFailure)
		caddy.RegisterEventHook("caddydb-cert-obtained", onDemandCertObtained)
	}
}

//OnDemandCertFailure Called when Caddy fails to obtain a certificate for a given host
func onDemandCertFailure(eventType caddy.EventName, eventInfo interface{}) error {
	if eventType != caddy.OnDemandCertFailureEvent {
		// Only listen to the event we are interested in
		return nil
	}

	// Interface containing data about a failed on demand certificate
	type CertFailureData struct {
		Name   string
		Reason error
	}

	data := eventInfo.(CertFailureData)
	go recordCertificateStatus(data.Name, "FAILED", data.Reason)
	return nil
}

//OnDemandCertObtained Called when Caddy obtains a certificate for a given host
func onDemandCertObtained(eventType caddy.EventName, eventInfo interface{}) error {
	if eventType != caddy.OnDemandCertObtainedEvent {
		// Only listen to the event we are interested in
		return nil
	}

	go recordCertificateStatus(eventInfo.(string), "LIVE", nil)
	return nil
}
