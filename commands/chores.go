package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sitzmaa/homebot/storage"
)

// HandleChore processes chore commands in #chores
func HandleChore(s *discordgo.Session, m *discordgo.MessageCreate) {
	parts := strings.Fields(m.Content)
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "!add":
		if len(parts) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: !add <chore description>")
			return
		}
		desc := strings.Join(parts[1:], " ")
		id, err := storage.AddChore(desc)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error adding chore: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added chore #%d: %s", id, desc))

	case "!chores":
		chores, err := storage.ListChores()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error listing chores: "+err.Error())
			return
		}
		if len(chores) == 0 {
			s.ChannelMessageSend(m.ChannelID, "No pending chores.")
			return
		}
		msg := "**Pending Chores:**\n"
		for _, c := range chores {
			status := "pending"
			if !c.CompletedAt.IsZero() {
				status = fmt.Sprintf("done by %s at %s", c.CompletedBy, c.CompletedAt.Format("2006-01-02 15:04"))
			}
			msg += fmt.Sprintf("#%d: %s — %s\n", c.ID, c.Description, status)
			for _, sub := range c.SubChores {
				subStatus := "pending"
				if !sub.CompletedAt.IsZero() {
					subStatus = fmt.Sprintf("done by %s at %s", sub.CompletedBy, sub.CompletedAt.Format("2006-01-02 15:04"))
				}
				msg += fmt.Sprintf("  - [%d.%d] %s — %s\n", c.ID, sub.ID, sub.Description, subStatus)
			}
		}
		s.ChannelMessageSend(m.ChannelID, msg)

	case "!subchore":
		if len(parts) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Usage: !subchore <choreID> <subchore description>")
			return
		}
		parent := parts[1]
		desc := strings.Join(parts[2:], " ")
		id, err := storage.AddSubChore(parent, desc)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error adding subchore: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added subchore #%s.%d: %s", parent, id, desc))

	case "!done":
		if len(parts) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: !done <choreID[.subID]>")
			return
		}
		target := parts[1]
		user := m.Author.Username
		err := storage.CompleteChore(target, user)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error marking done: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Marked %s as done by %s", target, user))
	}
}

