package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

var (
	Token     string = os.Getenv("TOKEN")
	ChannelID string = os.Getenv("CHAT")
	GuildID   string = os.Getenv("GUILD")
)

func main() {

	dg, err := discordgo.New("Bot " + Token)

	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	err = dg.Open()

	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	dg.AddHandler(messageCreate)

	fmt.Println("bot is now running, press CTRL-C to exit.")

	vc, err := connectToVoiceChannel(dg, GuildID, ChannelID)
	defer vc.Disconnect()

	if err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(10 * time.Second)
		close(vc.OpusRecv)
		vc.Close()
	}()

	handleVoice(vc.OpusRecv)

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
}

func connectToVoiceChannel(s *discordgo.Session, guildID, channelID string) (*discordgo.VoiceConnection, error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, false)
	if err != nil {
		return nil, err
	}

	return vc, nil
}

func handleVoice(c chan *discordgo.Packet) {
	files := make(map[uint32]media.Writer)
	for p := range c {
		file, ok := files[p.SSRC]
		if !ok {
			var err error
			file, err = oggwriter.New(fmt.Sprintf("%d.ogg", p.SSRC), 48000, 2)
			if err != nil {
				fmt.Printf("failed to create file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
				return
			}
			files[p.SSRC] = file
		}

		rtp := createPionRTPPacket(p)
		err := file.WriteRTP(rtp)
		if err != nil {
			fmt.Printf("failed to write to file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
		}
	}

	for _, f := range files {
		f.Close()
	}
}

func createPionRTPPacket(p *discordgo.Packet) *rtp.Packet {
	return &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    0x78,
			SequenceNumber: p.Sequence,
			Timestamp:      p.Timestamp,
			SSRC:           p.SSRC,
		},
		Payload: p.Opus,
	}
}
