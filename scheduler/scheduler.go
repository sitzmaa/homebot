package scheduler

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sitzmaa/homebot/storage"
)

// Start runs periodic tasks: sending reminders and pruning chores
func Start(s *discordgo.Session) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for now := range ticker.C {
		// Send due reminders
		rems, _ := storage.GetDueReminders(now)
		for _, r := range rems {
			s.ChannelMessageSend(r.ChannelID, r.Message)
			storage.UpdateReminderNext(r.ID)
		}
		// Prune chores completed >72h ago
		storage.PruneChores()
	}
}