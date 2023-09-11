package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	// "github.com/pion/rtp"
	// "github.com/pion/webrtc/v3/pkg/media"
	// "github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

var (
	Token     string
	ChannelID string
	GuildID   string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&GuildID, "g", "", "Guild in which voice channel exists")
	flag.StringVar(&ChannelID, "c", "", "Voice channel to connect to")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)

	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	err = dg.Open()

	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	dg.AddHandler(messageCreate)

	fmt.Println("bot is now running, press CTRL-C to exit.")

	vc, err := connectToVoiceChannel(dg, GuildID, ChannelID)
	if err != nil {
		panic(err)
	}

	// on CTRL-C, close connection
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("closing connection")
		vc.Disconnect()
		dg.Close()
		os.Exit(0)
	}()

	select {}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore own messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.ToLower(m.Content) == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong")
	}

	if strings.ToLower(m.Content) == "tic" {
		s.ChannelMessageSend(m.ChannelID, "tac")
	}
}

func connectToVoiceChannel(s *discordgo.Session, guildID, channelID string) (*discordgo.VoiceConnection, error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, false)
	if err != nil {
		return nil, err
	}

	return vc, nil
}
