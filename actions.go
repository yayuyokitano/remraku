package main

import (
	"context"
	"fmt"
	"math/rand"
)

func addXP(guildID string, userID string, nickname string, avatar string) (err error) {
	toAdd := rand.Int63n(11) + 15
	fmt.Println(nickname)
	fmt.Println(avatar)
	_, err = pool.Exec(context.Background(), "INSERT INTO guildXP (guildID, userID, nickname, avatar, xp) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (guildID, userID) DO UPDATE SET xp = guildXP.xp + EXCLUDED.xp, nickname = EXCLUDED.nickname, avatar = EXCLUDED.avatar", guildID, userID, nickname, avatar, toAdd)
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
