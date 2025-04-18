package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sitzmaa/homebot/storage"
)

// HandleTask processes task commands in #tasks
func HandleTask(s *discordgo.Session, m *discordgo.MessageCreate) {
	parts := strings.Fields(m.Content)
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "!todo":
		if len(parts) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: !todo <task description>")
			return
		}
		desc := strings.Join(parts[1:], " ")
		id, err := storage.AddTask(desc)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error adding task: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added task #%d: %s", id, desc))

	case "!tasks":
		tasks, err := storage.ListTasks()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error listing tasks: "+err.Error())
			return
		}
		if len(tasks) == 0 {
			s.ChannelMessageSend(m.ChannelID, "No pending tasks.")
			return
		}
		msg := "**Pending Tasks:**\n"
		for _, t := range tasks {
			msg += fmt.Sprintf("#%d: %s\n", t.ID, t.Description)
		}
		s.ChannelMessageSend(m.ChannelID, msg)

	case "!done":
		if len(parts) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: !done <taskID>")
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid task ID")
			return
		}
		err = storage.RemoveTask(id)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error removing task: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Removed task #%d", id))
	}
}