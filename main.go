package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/errorreporting"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4/pgxpool"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

var errorClient *errorreporting.Client
var datastoreClient *datastore.Client
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

	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)
	dg.AddHandler(guildDelete)

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

	datastoreClient, err = datastore.NewClient(ctx, GCP_PROJECT_ID)
	if err != nil {
		reportError(err)
		return
	}
	defer datastoreClient.Close()

	initRedis()

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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.Content == "remraku!ping" {
		s.ChannelMessageSend(m.ChannelID, "pong!")
	}

	if m.Content == "remraku!crash" {
		panic("test")
	}

	userHasPosted, err := checkUserHasPosted(m.Author.ID, m.GuildID)
	if err != nil {
		fmt.Println("error checking if user has posted:", err)
		return
	}

	if !userHasPosted {
		err = addXP(m.GuildID, m.Author.ID)
	}

	if m.Content == "remraku!test" {
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
