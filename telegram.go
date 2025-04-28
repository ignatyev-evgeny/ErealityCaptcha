package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"mime/multipart"
	"net/http"
	"time"
)

func sendToTelegram(token string, chatID int64, imgData []byte) (int, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", token)
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("photo", "captcha.png")
	if err != nil {
		return 0, err
	}
	_, err = part.Write(imgData)
	if err != nil {
		return 0, err
	}
	err = writer.WriteField("chat_id", fmt.Sprintf("%d", chatID))
	if err != nil {
		return 0, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("ошибка отправки: %s", resp.Status)
	}

	// Читаем ответ и парсим message_id
	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("ошибка парсинга ответа Telegram: %w", err)
	}

	return result.Result.MessageID, nil
}

func waitForCaptchaAnswer(botToken string, chatID int64, captchaMessageID int) (string, error) {
	log.Println("⏳ Ожидание ответа на капчу...")

	offset := 0

	for {
		time.Sleep(2 * time.Second)

		url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?timeout=5&offset=%d", botToken, offset)
		resp, err := http.Get(url)
		if err != nil {
			return "", fmt.Errorf("ошибка запроса к Telegram: %w", err)
		}
		defer resp.Body.Close()

		var updates struct {
			Ok     bool `json:"ok"`
			Result []struct {
				UpdateID int `json:"update_id"`
				Message  struct {
					MessageID int    `json:"message_id"`
					Text      string `json:"text"`
					Chat      struct {
						ID int64 `json:"id"`
					} `json:"chat"`
					ReplyToMessage *struct {
						MessageID int `json:"message_id"`
					} `json:"reply_to_message"`
				} `json:"message"`
			} `json:"result"`
		}

		err = json.NewDecoder(resp.Body).Decode(&updates)
		if err != nil {
			return "", fmt.Errorf("ошибка парсинга getUpdates: %w", err)
		}

		for _, update := range updates.Result {
			offset = update.UpdateID + 1

			// Проверяем что это ответ именно на капчу
			if update.Message.ReplyToMessage != nil &&
				update.Message.ReplyToMessage.MessageID == captchaMessageID &&
				update.Message.Chat.ID == chatID {
				return update.Message.Text, nil
			}
		}
	}
}

func listenForCommands(ctx context.Context, chatId int64) {
	log.Println("🖥 Старт прослушки команд Telegram...")

	offset := 0

	for {
		select {
		case <-ctx.Done():
			log.Println("⛔ Остановка прослушки команд.")
			return
		default:
			// Небольшая задержка между запросами
			time.Sleep(2 * time.Second)

			url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?timeout=5&offset=%d", BotToken, offset)
			resp, err := http.Get(url)
			if err != nil {
				log.Println("❌ Ошибка запроса к Telegram:", err)
				continue
			}
			defer resp.Body.Close()

			var updates struct {
				Ok     bool `json:"ok"`
				Result []struct {
					UpdateID int `json:"update_id"`
					Message  struct {
						MessageID int    `json:"message_id"`
						Text      string `json:"text"`
						Chat      struct {
							ID int64 `json:"id"`
						} `json:"chat"`
					} `json:"message"`
				} `json:"result"`
			}

			err = json.NewDecoder(resp.Body).Decode(&updates)
			if err != nil {
				log.Println("❌ Ошибка парсинга getUpdates:", err)
				continue
			}

			for _, update := range updates.Result {
				offset = update.UpdateID + 1

				if update.Message.Chat.ID == chatId {
					text := update.Message.Text

					if text == "/screen" {
						// (Твой уже существующий код обработки /screen)
						log.Println("📸 Получена команда /screen. Делаем скриншот!")

						var screenshot []byte
						err := chromedp.Run(ctx,
							chromedp.FullScreenshot(&screenshot, 90),
						)
						if err != nil {
							log.Println("❌ Ошибка создания скриншота:", err)
							continue
						}

						_, err = sendToTelegram(BotToken, chatId, screenshot)
						if err != nil {
							log.Println("❌ Ошибка отправки скриншота:", err)
							continue
						}

						log.Println("✅ Скриншот отправлен в Telegram!")
					}

					if text == "/click" {

						log.Println("🖱 Получена команда /click:")

						err := chromedp.Run(ctx,
							chromedp.MouseClickXY(650, 360),
						)
						if err != nil {
							log.Println("❌ Ошибка клика:", err)
							continue
						}

						log.Println("✅ Клик выполнен успешно!")
					}
				}
			}
		}
	}
}
