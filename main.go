package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func sendStartMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	var msg tgbotapi.MessageConfig = tgbotapi.NewMessage(update.Message.Chat.ID, "Привет, я - бот файлохранитель")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Покажи мне файлы которые лежат в моей папке")))
	msg.ReplyToMessageID = update.Message.MessageID
	bot.Send(msg)
}
func sendMyFiles(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	files, err := ioutil.ReadDir("Resource/" + update.Message.From.UserName)
	if err != nil {
		var msg tgbotapi.MessageConfig = tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет файлов")
		bot.Send(msg)
		sendStartMessage(bot, update)
	}
	var msg tgbotapi.MessageConfig = tgbotapi.NewMessage(update.Message.Chat.ID, "Список ваших файлов:")
	var Buttons []tgbotapi.KeyboardButton
	for _, file := range files {
		Buttons = append(Buttons, tgbotapi.NewKeyboardButton(file.Name()))
	}
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(Buttons)
	bot.Send(msg)
}
func sendFile(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	files, _ := ioutil.ReadDir("Resource/" + update.Message.From.UserName)
	for _, file := range files {
		if file.Name() == update.Message.Text {
			filePath := "Resource/" + update.Message.From.UserName + "/" + file.Name()
			fileBytes, _ := ioutil.ReadFile(filePath)
			msg := tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FileBytes{
				Name:  file.Name(),
				Bytes: fileBytes,
			})
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return
		}
	}
	var msg tgbotapi.MessageConfig = tgbotapi.NewMessage(update.Message.Chat.ID, "Файл не найден")
	bot.Send(msg)
	sendStartMessage(bot, update)
}
func getFile(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	fileID := update.Message.Document.FileID
	fileName := update.Message.Document.FileName

	fileURL, _ := bot.GetFileDirectURL(fileID)

	response, _ := http.Get(fileURL)
	defer response.Body.Close()

	_ = os.MkdirAll("Resource/"+update.Message.From.UserName, 0755)
	filePath := "Resource/" + update.Message.From.UserName + "/" + fileName
	file, _ := os.Create(filePath)
	_, _ = io.Copy(file, response.Body)
	sendStartMessage(bot, update)

}
func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}
	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Text == "/start" {
			sendStartMessage(bot, update)
		} else if update.Message.Text == "Покажи мне файлы которые лежат в моей папке" {
			sendMyFiles(bot, update)
		} else if update.Message.Document == nil {
			sendFile(bot, update)
		} else if update.Message.Document != nil {
			getFile(bot, update)
		}
	}
}
