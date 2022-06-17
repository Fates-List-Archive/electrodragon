package main

import (
	"errors"
	"os"
	"regexp"

	"github.com/MetroReviews/metro-integrase/types"
	log "github.com/sirupsen/logrus"
)

var regex *regexp.Regexp

func init() {
	var err error
	regex, err = regexp.Compile("[^a-zA-Z0-9]")

	if err != nil {
		panic(err)
	}
}

func addBot(bot *types.Bot) error {
	prefix := bot.Prefix

	if prefix == "" {
		prefix = "/"
	}

	invite := bot.Invite

	if invite == "" {
		invite = "https://discord.com/oauth2/authorize?client_id=" + bot.BotID + "&permissions=0&scope=bot%20applications.commands"
	}

	return nil
}

// Dummy adapter backend
type DummyAdapter struct {
}

func (adp DummyAdapter) GetConfig() types.ListConfig {
	return types.ListConfig{
		SecretKey:   os.Getenv("SECRET_KEY"),
		ListID:      os.Getenv("LIST_ID"),
		RequestLogs: true,
		StartupLogs: true,
		BindAddr:    ":1800",
		DomainName:  "",
	}
}

func (adp DummyAdapter) ClaimBot(bot *types.Bot) error {
	log.Info("Called ClaimBot")
	if bot == nil {
		return errors.New("bot is nil")
	}

	err := addBot(bot)

	if err != nil {
		return err
	}

	// TODO: Claim bot
	return nil
}

func (adp DummyAdapter) UnclaimBot(bot *types.Bot) error {
	log.Info("Called UnclaimBot")
	if bot == nil {
		return errors.New("bot is nil")
	}

	err := addBot(bot)

	if err != nil {
		return err
	}

	// TODO: Claim bot

	return nil
}

func (adp DummyAdapter) ApproveBot(bot *types.Bot) error {
	log.Info("Called ApproveBot")
	if bot == nil {
		return errors.New("bot is nil")
	}

	// TODO: Delete and readd it and approve bot
	return nil
}

func (adp DummyAdapter) DenyBot(bot *types.Bot) error {
	log.Info("Called DenyBot")
	if bot == nil {
		return errors.New("bot is nil")
	}

	// TODO: Delete and readd it and remove bot

	return nil
}

func (adp DummyAdapter) DataDelete(id string) error {
	return nil
}

func (adp DummyAdapter) DataRequest(id string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"id": id,
	}, nil
}
