package caddydb

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var session *mgo.Session

func connect() error {
	fmt.Println("[caddydb] Connecting to MongoDB...")
	newSession, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{"localhost"},
		Username: "caddy01",
		Password: "8Mj9GGb41eKE46AGCf44xUNqFfbjb1fQ897R84Lu3v5SA69W5L9sL2Gf3LSpw1651P97cerYn5dSXV7mneBEnEaC77Rjbkyem4ahXzPK39w23LSgX5T1VvkJt8a4S5gY",
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
