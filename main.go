package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/bwmarrin/discordgo"
	"github.com/sitzmaa/homebot/commands"
	"github.com/sitzmaa/homebot/scheduler"
	"github.com/sitzmaa/homebot/storage"
)

func main() {
	// Load local .env for development (ignored in production)
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN not set in environment")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	// Register handler
	dg.AddHandler(messageCreate)

	// Open connection
	if err := dg.Open(); err != nil {
		log.Fatalf("error opening connection: %v", err)
	}
	defer dg.Close()

	// Initialize storage (SQLite DB)
	if err := storage.Init("data.db"); err != nil {
		log.Fatalf("storage init failed: %v", err)
	}
	defer storage.Close()

	// Start background scheduler (reminders, pruning)
	go scheduler.Start(dg)

	log.Println("HouseBot is now running. Press CTRL-C to exit.")

	// Wait for termination signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages
	if m.Author.Bot {
		return
	}

	// Determine channel by name
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, _ = s.Channel(m.ChannelID)
	}
	name := ch.Name

	switch name {
	case "chores":
		commands.HandleChore(s, m)
	case "tasks":
		commands.HandleTask(s, m)
	case "shopping-list":
		// Placeholder for future shopping-list commands
	default:
		// Global reminder command
		if strings.HasPrefix(m.Content, "!reminder") {
			commands.HandleReminder(s, m)
		}
	}
}

