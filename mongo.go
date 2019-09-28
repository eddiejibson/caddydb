package caddydb

import (
	"fmt"
	"time"
	"encoding/json"
    	"os"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type Configuration struct {
    Password    string
}

var session *mgo.Session

func connect() error {
	file, _ := os.Open("/home/caddy/conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	errf := decoder.Decode(&configuration)
	if err != nil {
	  fmt.Println("error:", errf)
	}
	fmt.Println("[caddydb] Connecting to MongoDB...")
	newSession, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{"localhost"},
		Username: "caddy01",
		Password: configuration.Password,
		Database: "caddy",
	})

	if err != nil {
		fmt.Println("[caddydb] Unable to connect to database:", err)
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
