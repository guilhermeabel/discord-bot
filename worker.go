package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

func worker(activeGuild *ActiveGuild, guildId, channelId string) error {
	fmt.Printf("[%s] Starting worker\n", activeGuild.Name)
	defer cleanUpGuildWorker(activeGuild)

	for {
		media, ok := <-activeGuild.MediaChan

		if !ok {
			fmt.Println("empty loop")

			break
		}

		fmt.Println("Entered loop")
		fmt.Println("Media received from channel: ", media.FilePath)
		fmt.Printf("[%s] Worker joined audio\n", activeGuild.Name)

		voice, ok := <-activeGuild.VoiceConn

		if !ok {
			fmt.Println("empty conn")

			break
		}

		if !voice.Ready {
			voice.Disconnect()
			log.Printf("[%s] VoiceConnection no longer in ready state", activeGuild.Name)
			// voice, err = bot.ChannelVoiceJoin(guildId, channelId, false, true)
			// if err != nil {
			// 	return err
			// }
		}
		_ = voice.Speaking(true)

		if !activeGuild.UserActions.Stopped {
			play(voice, media, activeGuild)
		}
		_ = os.Remove(media.FilePath)
		if len(activeGuild.MediaChan) == 0 {
			break
		}
		log.Printf("[%s] There are currently %d medias in the queue", activeGuild.Name, activeGuild.MediaQueueSize())
		// Wait a bit before playing the next song
		time.Sleep(500 * time.Millisecond)
		_ = voice.Speaking(false)
		voice.Disconnect()
	}
	fmt.Println("passed loop")

	return nil
}

func cleanUpGuildWorker(activeGuild *ActiveGuild) {
	log.Printf("[%s] Cleaning up before destroying worker", activeGuild.Name)
	activeGuild.StopStreaming()
	log.Printf("[%s] Cleaned up all channels successfully", activeGuild.Name)
}

func play(voice *discordgo.VoiceConnection, media *Media, activeGuild *ActiveGuild) {
	options := StdEncodeOptions
	options.BufferedFrames = 100
	options.FrameDuration = 20
	options.CompressionLevel = 5
	options.Bitrate = 96

	encodeSession, err := EncodeFile(media.FilePath, options)
	if err != nil {
		log.Printf("[%s] Failed to create encoding session for \"%s\": %s", activeGuild.Name, media.FilePath, err.Error())
		return
	}
	defer encodeSession.Cleanup()

	time.Sleep(500 * time.Millisecond)

	log.Println("started playing")

	done := make(chan error)
	NewStream(encodeSession, voice, done)

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			log.Printf("[%s] Error occurred during stream for \"%s\": %s", activeGuild.Name, media.FilePath, err.Error())
			return
		}
	case <-activeGuild.UserActions.SkipChan:
		log.Printf("[%s] Skipping \"%s\"", activeGuild.Name, media.FilePath)
		_ = encodeSession.Stop()
	case <-activeGuild.UserActions.StopChan:
		log.Printf("[%s] Stopping", activeGuild.Name)
		_ = encodeSession.Stop()
	}
	return
}
