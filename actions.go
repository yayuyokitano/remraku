package main

import (
	"context"
	"log"
	"math/rand"

	"cloud.google.com/go/datastore"
)

type ServerXP map[string]int64

func (d *ServerXP) Load(props []datastore.Property) error {
	// Note: you might want to clear current values from the map or create a new map
	for _, p := range props {
		(*d)[p.Name] = p.Value.(int64)
	}
	return nil
}

func (d *ServerXP) Save() (props []datastore.Property, err error) {
	for k, v := range *d {
		props = append(props, datastore.Property{Name: k, Value: v})
	}
	return
}

func addXP(serverID string, userID string) (err error) {
	ctx := context.Background()
	tx, err := datastoreClient.NewTransaction(ctx)
	if err != nil {
		log.Printf("Could not create transaction: %v", err)
		return
	}

	serverXP := make(ServerXP)

	err = tx.Get(datastore.NameKey("ServerXP", serverID, nil), &serverXP)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			serverXP = make(ServerXP)
		} else {
			tx.Rollback()
			log.Printf("Could not get serverXP: %v", err)
			return
		}
	}

	toAdd := rand.Int63n(11) + 15
	serverXP[userID] += toAdd

	_, err = tx.Put(datastore.NameKey("ServerXP", serverID, nil), &serverXP)
	if err != nil {
		tx.Rollback()
		log.Printf("Could not put serverXP: %v", err)
		return
	}

	_, err = tx.Commit()
	if err != nil {
		log.Printf("Could not commit transaction: %v", err)
	}

	return
}

func addGuild(guildID string) (err error) {
	ctx := context.Background()
	_, err = pool.Exec(ctx, "INSERT INTO guilds (guildID) VALUES ($1)", guildID)
	return
}

func removeGuild(guildID string) (err error) {
	ctx := context.Background()
	_, err = pool.Exec(ctx, "DELETE FROM guilds WHERE guildID = $1", guildID)
	return
}
