package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/bwmarrin/discordgo"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

func main() {

	fmt.Fprintln(os.Stderr, "hello world")

	token, err := getToken()
	if err != nil {
		fmt.Println("error getting Discord token,", err)
		time.Sleep(time.Second * 60)
		main()
		return
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		time.Sleep(time.Second * 60)
		main()
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		time.Sleep(time.Second * 60)
		main()
		return
	}

	defer dg.Close()

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
	if m.Author.ID == s.State.User.ID {
		return
	}

	userHasPosted, err := checkUserHasPosted(m.Author.ID, m.GuildID)
	if err != nil {
		fmt.Println("error checking if user has posted:", err)
		return
	}

	if m.Content == "remraku!test" {
		if userHasPosted {
			s.ChannelMessageSend(m.ChannelID, "spammer! (but we do make CD work)")
		} else {
			s.ChannelMessageSend(m.ChannelID, "hi!")
		}
	}
}

func getToken() (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/rem-970606/secrets/DISCORD_TOKEN/versions/latest",
	}
	resp, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}

	return string(resp.Payload.Data), nil

}
