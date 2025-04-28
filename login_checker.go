package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

func checkLogin(ctx context.Context, login, password string) {
	log.Println("🛡 Старт проверки наличия формы логина каждые 5 секунд...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("⛔ Остановка проверки авторизации (контекст завершён).")
			return
		case <-ticker.C:
			var loginFormVisible bool
			err := chromedp.Run(ctx,
				chromedp.Evaluate(`document.querySelector('input[name="login"]') !== null`, &loginFormVisible),
			)
			if err != nil {
				log.Println("❌ Ошибка проверки формы логина:", err)
				continue
			}

			if loginFormVisible {
				log.Println("⚠️ Форма логина найдена! Пытаемся авторизоваться снова...")

				err := chromedp.Run(ctx,
					chromedp.SendKeys(`input[name="login"]`, login, chromedp.ByQuery),
					chromedp.SendKeys(`input[name="pass"]`, password, chromedp.ByQuery),
					chromedp.Click(`input.btn_enter`, chromedp.ByQuery),
					chromedp.Sleep(2*time.Second),
				)
				if err != nil {
					log.Println("❌ Ошибка при повторной авторизации:", err)
					continue
				}

				log.Println("✅ Повторная авторизация прошла успешно!")

				clickX := 650
				clickY := 365

				err = chromedp.Run(ctx,
					chromedp.MouseClickXY(float64(clickX), float64(clickY)),
				)
				if err != nil {
					log.Fatal("❌ Ошибка при клике:", err)
				}

			} else {
				log.Println("✅ Форма логина не найдена. Всё в порядке.")
			}
		}
	}
}
