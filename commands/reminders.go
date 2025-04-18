package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sitzmaa/homebot/storage"
)

// HandleReminder processes !reminder commands in any channel
func HandleReminder(s *discordgo.Session, m *discordgo.MessageCreate) {
	parts := strings.Fields(m.Content)
	if len(parts) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !reminder <daily|weekly|monthly> <message>")
		return
	}
	freq := strings.ToLower(parts[1])
	msg := strings.Join(parts[2:], " ")
	rem, err := storage.AddReminder(freq, msg, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error scheduling reminder: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Scheduled reminder #%d: %s (%s)", rem.ID, rem.Message, rem.Frequency))
}