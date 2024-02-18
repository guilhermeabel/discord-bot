package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	// https://github.com/TwiN/discord-music-bot
)

var (
	Token          string = os.Getenv("TOKEN")
	ChannelID      string = os.Getenv("CHAT")
	GuildID        string = os.Getenv("GUILD")
	guilds                = make(map[string]*ActiveGuild)
	youtubeService *Service
	guildsMutex    = sync.RWMutex{}

	guildNames = make(map[string]string)
)

func main() {
	youtubeService = NewService(480)

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

	// vc, err := connectToVoiceChannel(dg, GuildID, ChannelID)
	// defer vc.Disconnect()

	if err != nil {
		panic(err)
	}

	app := &applicationServer{}
	err = app.server().ListenAndServe()

	if err != nil {
		panic(err)
	}
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

	if strings.ToLower(m.Content) == "tico" {
		s.ChannelMessageSend(m.ChannelID, "teco")
	}

	if strings.ToLower(m.Content) == "teque" {
		s.ChannelMessageSend(m.ChannelID, "teque")
	}

	if strings.ToLower(m.Content) == "play" {
		activeGuild := guilds[m.GuildID]
		HandleYoutubeCommand(s, activeGuild, m, "https://www.youtube.com/watch?v=0iV23FhyJ_A")
	}
}

func connectToVoiceChannel(s *discordgo.Session, guildID, channelID string) (*discordgo.VoiceConnection, error) {
	fmt.Println("connecting to voice channel")
	fmt.Println("guildID: ", guildID)
	fmt.Println("channelID: ", channelID)

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, false)
	if err != nil {
		fmt.Println("error joining voice channel,", err)
		return nil, err
	}
	fmt.Println("printed,")

	return vc, nil
}

func HandleYoutubeCommand(bot *discordgo.Session, activeGuild *ActiveGuild, message *discordgo.MessageCreate, query string) {
	if activeGuild == nil {
		activeGuild = NewActiveGuild(GetGuildNameByID(bot, message.GuildID))
		guildsMutex.Lock()
		guilds[message.GuildID] = activeGuild
		guildsMutex.Unlock()
	}

	voiceChannelId, err := GetVoiceChannelWhereMessageAuthorIs(bot, message)

	if err != nil {
		log.Printf("[%s] Failed to get voice channel: %s", activeGuild.Name, err.Error())
		_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("❌ Unable to get voice channel: %s", err.Error()))
		return
	}

	media, err := youtubeService.SearchAndDownload(query)
	if err != nil {
		log.Printf("[%s] Failed to search and download media: %s", activeGuild.Name, err.Error())
		_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("unable to search and download media: %s", err.Error()))
		return
	}

	createNewWorker := false
	if !activeGuild.IsStreaming() {
		log.Printf("[%s] Preparing for streaming", activeGuild.Name)
		activeGuild.PrepareForStreaming(10)

		createNewWorker = true
	}

	activeGuild.EnqueueMedia(media)
	log.Printf("[%s] Enqueued: %s", activeGuild.Name, media.Title)

	go func() {
		voice, err := connectToVoiceChannel(bot, message.GuildID, voiceChannelId)
		if err != nil {
			fmt.Println("error connecting to voice channel,", err)
		}

		activeGuild.SetVoiceChannel(voice)
	}()

	if createNewWorker {
		go func() {
			fmt.Println("creating worker")
			err = worker(activeGuild, message.GuildID, voiceChannelId)
			if err != nil {
				log.Printf("[%s] Failed to start worker: %s", activeGuild.Name, err.Error())
				// _, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("unable to start voice worker: %s", err.Error()))
				_ = os.Remove(media.FilePath)
				return
			}
		}()
	}

}

func GetVoiceChannelWhereMessageAuthorIs(bot *discordgo.Session, message *discordgo.MessageCreate) (string, error) {
	guild, err := bot.State.Guild(message.GuildID)
	if err != nil {
		return "", err
	}
	for _, voiceState := range guild.VoiceStates {
		if voiceState.UserID == message.Author.ID {
			return voiceState.ChannelID, nil
		}
	}
	return "", fmt.Errorf("you are not in a voice channel")
}

func GetGuildNameByID(bot *discordgo.Session, guildID string) string {
	guildName, ok := guildNames[guildID]
	if !ok {
		guild, err := bot.Guild(guildID)
		if err != nil {
			// Failed to get the guild? Whatever, we'll just use the guild id
			guildNames[guildID] = guildID
			return guildID
		}
		guildNames[guildID] = guild.Name
		return guild.Name
	}
	return guildName
}
