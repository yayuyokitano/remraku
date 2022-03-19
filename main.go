package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/errorreporting"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/bwmarrin/discordgo"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

var errorClient *errorreporting.Client
var pool *pgxpool.Pool

func main() {

	ctx := context.Background()

	token, err := getToken()
	if err != nil {
		fmt.Println("error getting Discord token,", err)
		time.Sleep(time.Second * 10)
		main()
		return
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		time.Sleep(time.Second * 10)
		main()
		return
	}

	databaseURL, err := getDatabaseURL()
	if err != nil {
		reportError(err)
		return
	}

	pool, err = pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		time.Sleep(time.Second * 10)
		main()
		return
	}
	defer dg.Close()

	errorClient, err = errorreporting.NewClient(ctx, GCP_PROJECT_ID, errorreporting.Config{
		ServiceName: "remraku",
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		fmt.Println("error creating log error client,", err)
		time.Sleep(time.Second * 10)
		main()
		return
	}
	defer errorClient.Close()

	initRedis()

	err = loadBlocklist()
	if err != nil {
		fmt.Println("error loading blocklist:", err)
		time.Sleep(time.Second * 10)
		main()
		return
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)
	dg.AddHandler(guildDelete)

	fmt.Println(".-------.        .-''-.  ,---.    ,---..-------.       ____    .--.   .--.    ___    _  ")
	fmt.Println("|  _ _   \\     .'_ _   \\ |    \\  /    ||  _ _   \\    .'  __ `. |  | _/  /   .'   |  | | ")
	fmt.Println("| ( ' )  |    / ( ` )   '|  ,  \\/  ,  || ( ' )  |   /   '  \\  \\| (`' ) /    |   .'  | | ")
	fmt.Println("|(_ レ_) /   . (_ o _)  ||  |\\_   /|  ||(_ム _) /   |___|  /  ||(_ ()_)     .'  '_  | | ")
	fmt.Println("| (_,_).' __ |  (_,_)___||  _( )_/ |  || (_,_).' __    _.-`   || (_,_)   __ '   ( \\.-.| ")
	fmt.Println("|  |\\ \\  |  |'  \\   .---.| (_ o _) |  ||  |\\ \\  |  |.'   _    ||  |\\ \\  |  |' (`. _` /| ")
	fmt.Println("|  | \\ `'   / \\  `-'    /|  (_,_)  |  ||  | \\ `'   /|  _( )_  ||  | \\ `'   /| (_ (絡)_) ")
	fmt.Println("|  |  \\    /   \\      /  |  |      |  ||  |  \\    / \\ (_ o _) /|  |  \\    /  \\ /  . \\ / ")
	fmt.Println("''-'   `'-'     `'-..-'  '--'      '--'''-'   `'-'   '.(_,_).' `--'   `'-'    ``-'`-''  ")

	fmt.Println("Remraku is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func getName(m *discordgo.MessageCreate) string {
	if m.Member.Nick == "" {
		return m.Author.Username
	}
	return m.Member.Nick
}

func getAvatar(m *discordgo.MessageCreate) string {
	if m.Member.Avatar != "" {
		return m.Member.Avatar
	}
	return m.Author.Avatar
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.Content == "remraku!ping" {
		s.ChannelMessageSend(m.ChannelID, "pong!")
	}

	userHasPosted, err := checkUserHasPosted(m.Author.ID, m.GuildID)
	if err != nil {
		fmt.Println("error checking if user has posted:", err)
		return
	}

	xpBlocked, err := checkBlocklist(m.ChannelID, "xpgain")
	if err != nil {
		fmt.Println("error checking blocklist:", err)
		return
	}

	if !userHasPosted && !xpBlocked {
		err = addXP(m.GuildID, m.Author.ID, getName(m), getAvatar(m))
	}
	if err != nil {
		fmt.Println("error adding xp:", err)
		return
	}

	if m.Content == "remraku!dump" && m.Author.ID == "196249128286552064" {
		s.ChannelMessageSend(m.ChannelID, "dumping...")
		keys, err := rdb.Do("KEYS", "*").Result()
		fmt.Println(keys)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("error dumping redis: %v", err))
		}
		r := ""
		for _, key := range keys.([]interface{}) {
			r += key.(string) + "\n"
		}

		s.ChannelMessageSend(m.ChannelID, "```\n"+r+"\n```")
	}

	if m.Content == "remraku!test" {
		if xpBlocked {
			s.ChannelMessageSend(m.ChannelID, "blocklisted!")
			return
		}
		if userHasPosted {
			s.ChannelMessageSend(m.ChannelID, "spammer! (but CD is working!)")
		} else {
			s.ChannelMessageSend(m.ChannelID, "hi!")
		}
	}
}

func guildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	if g.Guild.Unavailable {
		return
	}

	err := addGuild(g.Guild.ID)
	if err != nil {
		fmt.Println("error adding guild:", err)
		return
	}

}

func guildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	err := removeGuild(g.Guild.ID)
	if err != nil {
		fmt.Println("error removing guild:", err)
		return
	}
}

func getSecret(name string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/" + GCP_PROJECT_ID + "/secrets/" + name + "/versions/latest",
	}
	resp, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}

	return string(resp.Payload.Data), nil
}

func getToken() (string, error) {
	return getSecret("DISCORD_TOKEN")
}

func getDatabaseURL() (string, error) {
	return getSecret("DATABASE_PRIVATE_URL")
}

func reportError(err error) {
	errorClient.Report(errorreporting.Entry{
		Error: err,
	})
	log.Print(err)
}

type Blocked struct {
	GuildID   string `db:"guildid"`
	ChannelID string `db:"channelid"`
	Xpgain    bool   `db:"xpgain"`
}

func loadBlocklist() (err error) {
	ctx := context.Background()
	var blocklist []Blocked
	err = pgxscan.Select(ctx, pool, &blocklist, "SELECT * FROM channelblocklist")
	if err != nil {
		return err
	}
	fmt.Println(blocklist)
	for _, channel := range blocklist {
		fmt.Println(channel)
		modifyBlocklist(channel.GuildID, "xpgain", channel.Xpgain)
	}
	return
}
