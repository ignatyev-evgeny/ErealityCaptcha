package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func startBot(login, password string, chatId int64) {

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.WindowSize(1024, 1024),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Авторизация
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://ereality.ru/login/`),
		chromedp.WaitVisible(`input[name="login"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="login"]`, login, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="pass"]`, password, chromedp.ByQuery),
		chromedp.Click(`input.btn_enter`, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatal("❌ Ошибка при авторизации:", err)
	}

	log.Println("✅ Успешный вход в систему.")

	go checkLogin(ctx, login, password)

	go listenForCommands(ctx, chatId)

	go waitForCaptchaAndSolve(ctx, chatId)

	// Ожидание Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	log.Println("⏳ Ожидание сигнала завершения (Ctrl+C)...")
	<-sig

	log.Println("⛔ Сигнал завершения получен. Завершаем работу...")

	if ctxCancel != nil {
		ctxCancel()
		log.Println("✅ Контекст работы закрыт.")
	}

	if allocCancel != nil {
		allocCancel()
		log.Println("✅ Браузер Chrome закрыт.")
	}

	log.Println("✅ Всё корректно завершено.")
}
