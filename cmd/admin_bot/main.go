package main

import (
	"context"
	"github.com/damianino/dcw-shop-bot/config"
	"github.com/damianino/dcw-shop-bot/internal/admin_bot/chat_tree"
	"github.com/damianino/dcw-shop-bot/internal/repository"
	"github.com/damianino/dcw-shop-bot/pkg/mysql"
	"github.com/damianino/dcw-shop-bot/pkg/telegram_bot_framework"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	//ctx := context.Background()

	//go func() {
	//	<- ctx.Done()
	//
	//}()

	if err := godotenv.Load(".env"); err != nil {
		log.Println("failed to load env")
	}

	conf, err := config.GetConfig()
	if err != nil {
		log.Panicf("failed to parse config: %s", err.Error())
	}

	//dsn := (&url.URL{
	//	Scheme: "mysql",
	//	User:   url.User(conf.DB.Username), //url.UserPassword(conf.DB.Username, conf.DB.Password),
	//	Host:   net.JoinHostPort(conf.DB.Host, conf.DB.Port),
	//	Path:   conf.DB.Name,
	//}).String()
	//println(dsn)
	db, err := mysql.NewDB("root:@/dcw_shop_bot?parseTime=true")
	if err != nil {
		log.Println(err)
	}

	repo := repository.NewRepo(db)

	tgBotAPI, err := tgbotapi.NewBotAPI(conf.TelegramAdminBot.Token)
	if err != nil {
		log.Println(err)

		return
	}

	chatTree := chat_tree.InitChatTree(&chat_tree.Deps{repo})

	bot := telegram_bot_framework.NewBot(tgBotAPI, chatTree)

	bot.StartBot(context.Background())
}
