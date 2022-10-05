package main

import (
	"log"

	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/tg"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/report"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/spend"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/config"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/command"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/spending"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatal("config init failed:", err)
	}

	tgClient, err := tg.New(config)
	if err != nil {
		log.Fatal("tg client init failed:", err)
	}

	commandStorage := command.NewStorage()
	spendingStorage := spending.NewStorage()

	msgModel := messages.New()

	msgModel.SetDefaultCommand(commands.NotFoundCommand(tgClient))
	msgModel.AddCommand("/start", commands.Hello(tgClient))
	msgModel.AddCommand("/spend", spend.New(commandStorage, spendingStorage, tgClient))
	msgModel.AddCommand("/help", commands.Help(tgClient))
	msgModel.AddCommand("/report", report.New(spendingStorage, tgClient, commandStorage))

	tgClient.ListenUpdates(msgModel)
}
