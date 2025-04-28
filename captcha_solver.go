package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"time"
)

func waitForCaptchaAndSolve(ctx context.Context, chatId int64) {

	log.Println("🔎 Начинаем проверку появления капчи каждые 2 секунды...")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("⛔ Остановка проверки капчи из-за завершения контекста.")
			return

		case <-ticker.C:
			log.Println("🔄 Проверка наличия iframe карты...")

			var iframeFound bool
			err := chromedp.Run(ctx,
				chromedp.Evaluate(`document.querySelector('iframe[src*="map.php"]') !== null`, &iframeFound),
			)
			if err != nil {
				log.Println("❌ Ошибка при поиске iframe:", err)
				continue
			}

			if !iframeFound {
				log.Println("⚠️ iframe карты пока не найден.")
				continue
			}

			log.Println("✅ iframe найден. Проверяем капчу внутри...")

			// Проверка наличия капчи внутри iframe через простой JavaScript
			var captchaExists bool
			err = chromedp.Run(ctx,
				chromedp.Evaluate(`(() => {
					const iframe = document.querySelector('iframe[src*="map.php"]');
					if (!iframe) return false;
					try {
						return iframe.contentDocument.getElementById('captchaImage') !== null;
					} catch (e) {
						return false;
					}
				})()`, &captchaExists),
			)
			if err != nil {
				log.Println("❌ Ошибка проверки капчи внутри iframe:", err)
				continue
			}

			if captchaExists {
				log.Println("🛡 Капча найдена! Готовимся к решению!")

				imgData, err := captureCaptcha(ctx, 250, 85, 500, 350)
				if err != nil {
					log.Println("❌ Ошибка при захвате капчи:", err)
					return
				}

				var captchaMessageID int

				captchaMessageID, err = sendToTelegram("7783160085:AAGdcKa1aCYL3lwYJRyfzcR0eh2qrm3pspo", chatId, imgData)
				if err != nil {
					log.Println("❌ Ошибка при отправке капчи:", err)
					return
				}

				answer, err := waitForCaptchaAnswer("7783160085:AAGdcKa1aCYL3lwYJRyfzcR0eh2qrm3pspo", chatId, captchaMessageID)
				if err != nil {
					log.Println("❌ Ошибка получения ответа на капчу:", err)
					return
				}
				log.Println("✅ Ответ на капчу:", answer)

				err = clickCaptchaAnswer(ctx, answer)
				if err != nil {
					log.Println("❌ Ошибка клика по капче:", err)
				}

				clickX := 650
				clickY := 365

				err = chromedp.Run(ctx,
					chromedp.MouseClickXY(float64(clickX), float64(clickY)),
				)
				if err != nil {
					log.Fatal("❌ Ошибка при клике:", err)
				}

				go waitForCaptchaAndSolve(ctx, chatId)

				// Здесь будет решалка капчи
				return
			}

			log.Println("❌ Капча не найдена в iframe. Продолжаем проверку...")
		}
	}
}

func captureCaptcha(ctx context.Context, x, y, width, height int) ([]byte, error) {
	var fullScreenshot []byte
	err := chromedp.Run(ctx,
		chromedp.FullScreenshot(&fullScreenshot, 90),
	)
	if err != nil {
		return nil, err
	}

	img, err := jpeg.Decode(bytes.NewReader(fullScreenshot))
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования скрина: %w", err)
	}

	captcha := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(x, y, x+width, y+height))

	var buf bytes.Buffer
	err = png.Encode(&buf, captcha)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func clickCaptchaAnswer(ctx context.Context, answer string) error {
	// Карта цифра => координаты (X, Y)
	coords := map[rune][2]int{
		'7': {650, 360}, // Сапоги
		'6': {600, 360}, // Доспехи
		'5': {550, 360}, // Перчатки
		'4': {500, 360}, // Наручи
		'3': {440, 360}, // Шлемы
		'2': {380, 360}, // Кулоны
		'1': {330, 360}, // Кольца
	}

	if len(answer) != 2 {
		return fmt.Errorf("неправильный формат ответа на капчу: %s", answer)
	}

	for i, digit := range answer {
		pos, ok := coords[digit]
		if !ok {
			return fmt.Errorf("неизвестная цифра в ответе: %c", digit)
		}

		log.Printf("🖱 Кликаем по координатам (%d, %d) для цифры %c", pos[0], pos[1], digit)

		err := chromedp.Run(ctx,
			chromedp.MouseClickXY(float64(pos[0]), float64(pos[1])),
		)
		if err != nil {
			return fmt.Errorf("ошибка клика по координатам (%d, %d): %w", pos[0], pos[1], err)
		}

		if i == 0 {
			// После первого клика — подождать 0.5 секунды
			time.Sleep(500 * time.Millisecond)
		}
	}

	return nil
}
