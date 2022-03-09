package main

import (
	"context"
	"math/rand"
)

func addXP(guildID string, userID string) (err error) {
	toAdd := rand.Int63n(11) + 15
	_, err = pool.Exec(context.Background(), "INSERT INTO guildXP (guildID, userID, xp) VALUES ($1, $2, $3) ON CONFLICT (xplookup) DO UPDATE SET xp = xp + $3", guildID, userID, toAdd)
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
