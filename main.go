package main

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

var (
	allocCancel context.CancelFunc
	ctxCancel   context.CancelFunc
)

var (
	BotToken string
)

func main() {

	// Загружаем .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Ошибка загрузки .env")
	}

	// Get BOT_TOKEN from .env
	BotToken := os.Getenv("BOT_TOKEN")
	if BotToken == "" {
		log.Fatal("❌ Переменная BOT_TOKEN не определена в .env")
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("Ereality Captcha Resolve")

	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Login")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	chatIdEntry := widget.NewEntry()
	chatIdEntry.SetPlaceHolder("Chat ID")

	settings, err := loadSettings()
	if err == nil {
		loginEntry.SetText(settings.Login)
		passwordEntry.SetText(settings.Password)
		chatIdEntry.SetText(settings.ChatID)
	}

	startButton := widget.NewButton("Start bot", func() {
		login := loginEntry.Text
		password := passwordEntry.Text
		chatIdStr := chatIdEntry.Text

		// Сохраняем настройки
		err := saveSettings(Settings{
			Login:    login,
			Password: password,
			ChatID:   chatIdStr,
		})

		if err != nil {
			log.Println("❌ Ошибка сохранения настроек")
			return
		}

		chatId, err := strconv.ParseInt(chatIdStr, 10, 64)
		if err != nil {
			log.Println("❌ Ошибка преобразования Chat ID:", err)
			return
		}

		go startBot(login, password, chatId)
	})

	// Обработчик нажатия на крестик
	myWindow.SetCloseIntercept(func() {
		log.Println("⛔ Нажат крестик! Завершаем работу...")

		if ctxCancel != nil {
			ctxCancel()
			log.Println("✅ Контекст работы закрыт.")
		}

		if allocCancel != nil {
			allocCancel()
			log.Println("✅ Браузер Chrome закрыт.")
		}

		myWindow.Close()
	})

	myWindow.SetContent(container.NewVBox(
		loginEntry,
		passwordEntry,
		chatIdEntry,
		startButton,
	))

	myWindow.Resize(fyne.NewSize(300, 200))
	myWindow.ShowAndRun()
}
