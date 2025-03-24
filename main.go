package main

import (
	"fmt"
	"io"
	"kursach/client"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/telebot.v3"
)

func main() {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  "7791869623:AAHfmFPDKCoIvbRoHgO6_YKO32r_EtN1vTc",
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second}, // Тут посмотреть про поллинг и вебсокеты
	})
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		modelPath := filepath.Join("voices", "1.wav")
		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			return c.Send("Сначала отправьте голосовое сообщение для создания модели!")
		}

		modelBytes, err := os.ReadFile(modelPath)
		if err != nil {
			return c.Send("Ошибка чтения голосовой модели")
		}

		text := c.Text()
		if text == "" {
			return c.Send("Текст не может быть пустым")
		}

		audioBytes, err := client.SendAudio(text, modelBytes)
		if err != nil {
			return c.Send("Ошибка генерации аудио: " + err.Error())
		}

		file, _ := os.Create("alen.wav")

		_, err = file.Write(audioBytes.Result.ProcessedAudio)
		if err != nil {
			return err
		}

		// Отправка результата
		voice := &telebot.Voice{
			File:     telebot.FromDisk("alen.wav"),
			MIME:     "audio/ogg",
			Duration: 5,
		}

		return c.Send(voice)
	})

	bot.Handle(telebot.OnVoice, func(c telebot.Context) error {
		if err := os.MkdirAll("voices", 0755); err != nil {
			return c.Send("Ошибка создания директории")
		}

		voice := c.Message().Voice
		if voice == nil {
			return c.Send("Это не голосовое сообщение")
		}

		fileInfo, err := bot.FileByID(voice.FileID)
		if err != nil {
			return c.Send("Ошибка получения файла")
		}

		fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", bot.Token, fileInfo.FilePath)

		filePath := filepath.Join("voices", "1.wav")

		resp, err := http.Get(fileURL)
		if err != nil {
			return c.Send("Ошибка скачивания файла")
		}
		defer resp.Body.Close()

		outFile, err := os.Create(filePath)
		if err != nil {
			return c.Send("шибка создания файла")
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, resp.Body); err != nil {
			return c.Send("Ошибка сохранения файла")
		}

		return c.Send("Модель успешно сохранена")
	})

	log.Println("Бот запущен...")
	bot.Start()
}

func downloadFile(fileName, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("ошибка HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неверный статус код: %d", resp.StatusCode)
	}

	outFile, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("ошибка копирования данных: %w", err)
	}

	return nil
}
