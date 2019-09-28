package caddydb

import (
	"github.com/caddyserver/caddy"
	"fmt"
	"time"
	"encoding/json"
    	"os"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func init() {
	caddy.RegisterPlugin("caddydb", caddy.Plugin{
		ServerType: "http",
		Action:     Setup,
	})
	err := connect()

	// Only register event hooks if database connection succeeds
	if err == nil {
		caddy.RegisterEventHook("caddydb-cert-failure", onDemandCertFailure)
		caddy.RegisterEventHook("caddydb-cert-obtained", onDemandCertObtained)
	}
}

func Setup(c *caddy.Controller) error {
	return nil
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


type Config struct {
    Password    string
}

var session *mgo.Session

func connect() error {

	config := LoadConfiguration()
	fmt.Println(config)
	fmt.Println("[caddydb] Connecting to MongoDB...")
	newSession, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{"localhost"},
		Username: "caddy01",
		Password: config.Password,
		Database: "caddy",
	})

	if err != nil {
		fmt.Println("[caddydb] Unable to connect to database with password:", config.Password, err)
		return err
	}

	info, err := newSession.BuildInfo()

	if err != nil {
		fmt.Println("[caddydb] Connection was established but unable to fetch server info:", err)
		return err
	}

	session = newSession
	fmt.Println("[caddydb] Connection established to MongoDB server version", info.Version)

	return nil
}

func LoadConfiguration() Config {
    var config Config
    configFile, err := os.Open("/home/caddy/conf.json")
    defer configFile.Close()
    if err != nil {
        fmt.Println(err.Error())
    }
    jsonParser := json.NewDecoder(configFile)
    jsonParser.Decode(&config)
    return config
}

func recordCertificateStatus(domain string, status string, failureReason error) error {
	collection := session.DB("caddy").C("certificates")
	selector := bson.M{"domain": domain}
	newStatus := bson.M{"status": status, "lastUpdate": time.Now()}

	if failureReason != nil {
		newStatus["error"] = failureReason.Error()
	}

	update := bson.M{"$set": newStatus}
	_, err := collection.Upsert(selector, update)

	if err == nil {
		fmt.Printf("[caddydb] Saved certificate request for %s (new status: %s)\n", domain, status)
	} else {
		fmt.Printf("[caddydb] Unable to save certificate request for %s: %s\n", domain, err)
	}

	return err
}
