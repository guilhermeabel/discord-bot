package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// set environment variables
	config()

	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

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

	vc, err := connectToVoiceChannel(dg, os.Getenv("GUILD_ID"), os.Getenv("CHANNEL_ID"))
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
	if m.Author.ID == s.State.User.ID {
		// ignore own messages
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
