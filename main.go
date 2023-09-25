package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var triggers []string

func main() {
	botToken, err := loadToken()
	if err != nil {
		log.Fatalf("Помилка завантаження токена: %v\n", err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Помилка при підключенні до API: %v\n", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Помилка при отриманні оновлень: %v\n", err)
	}

	log.Println("Бот запущений та готовий до прийому команд...")

	// Перед початком виконання команд /save та /del оновлюємо список триггерів з файлу
	updateTriggerList()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				handleStartCommand(bot, update.Message)
			case "save":
				handleSaveCommand(bot, update.Message)
			case "del":
				handleDeleteCommand(bot, update.Message)
			case "list":
				sendTriggerList(bot, update.Message.Chat.ID)
			case "ping":
				handlePingCommand(bot, update.Message)
			case "help":
				handleHelpCommand(bot, update.Message)
			}
		} else {
			handleVideoMessage(bot, update.Message)
			sendTriggeredMessage(bot, update.Message.Chat.ID, update.Message.Text, update.Message.From.UserName)
		}
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	welcomeText := "Вітаємо! Цей бот допоможе вам зберігати та використовувати тригери. Використовуйте команди /save, /del, /list, /ping та /help для управління ботом."
	sendMessage(bot, message.Chat.ID, welcomeText)
}

func handleSaveCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	re := regexp.MustCompile(`/save\s+(.+)`)
	match := re.FindStringSubmatch(message.Text)
	if match == nil || message.ReplyToMessage == nil {
		return
	}
	trigger := match[1]

	// Перевіряємо, що тригер є унікальним
	if !isTriggerUnique(trigger) {
		sendMessage(bot, message.Chat.ID, "Цей тригер вже існує.")
		return
	}

	mediaType, mediaID := extractMediaID(message.ReplyToMessage)
	saveTrigger(trigger, message.ReplyToMessage.Text, mediaType, mediaID)
	sendMessage(bot, message.Chat.ID, "Тригер збережено успішно.")
}

func handleDeleteCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	re := regexp.MustCompile(`/del\s+(.+)`)
	match := re.FindStringSubmatch(message.Text)
	if match == nil {
		return
	}
	trigger := match[1]

	// Перевіряємо, що тригер існує
	if !isTriggerExists(trigger) {
		sendMessage(bot, message.Chat.ID, "Цього тригера не існує.")
		return
	}

	deleteTrigger(trigger)
	sendMessage(bot, message.Chat.ID, "Тригер видалено успішно.")
}

func handlePingCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	startTime := time.Now()
	msg := tgbotapi.NewMessage(message.Chat.ID, "Понг!")
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Помилка при відправці повідомлення: %v\n", err)
		return
	}

	elapsedTime := time.Since(startTime)
	responseMsg := fmt.Sprintf("Час відповіді бота: %v", elapsedTime)
	msg = tgbotapi.NewMessage(message.Chat.ID, responseMsg)
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("Помилка при відправці повідомлення: %v\n", err)
	}
}

func handleHelpCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	helpText := `
Доступні команди:
/start - почати використовувати бота
/save <тригер> - зберегти тригер
/del <тригер> - видалити тригер
/list - вивести список тригерів
/ping - перевірити доступність бота та його швидкість відповіді
/help - отримати цю довідку`

	sendMessage(bot, message.Chat.ID, helpText)
}

func handleVideoMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	if message.VideoNote != nil {
		re := regexp.MustCompile(`/save\s+(.+)`)
		match := re.FindStringSubmatch(message.Text)
		if match == nil {
			return
		}
		trigger := match[1]
		fileID := (*message.VideoNote).FileID
		saveTrigger(trigger, message.Text, "videonote", fileID)
	}
}

func extractMediaID(message *tgbotapi.Message) (string, string) {
	if message.Photo != nil && len(*message.Photo) > 0 {
		return "photo", (*message.Photo)[0].FileID
	}
	if message.Sticker != nil {
		return "sticker", message.Sticker.FileID
	}
	if message.Video != nil {
		return "video", message.Video.FileID
	}
	if message.Voice != nil {
		return "voice", message.Voice.FileID
	}
	if message.Audio != nil {
		return "audio", message.Audio.FileID
	}
	if message.Animation != nil {
		return "animation", message.Animation.FileID
	}
	if message.VideoNote != nil {
		return "videonote", message.VideoNote.FileID
	}
	return "text", ""
}

func saveTrigger(trigger, messageText, mediaType, mediaID string) {
	file, err := os.OpenFile("triggers.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Помилка при відкритті файлу: %v\n", err)
		return
	}
	defer file.Close()

	data := fmt.Sprintf("%s|%s|%s|%s\n", trigger, messageText, mediaType, mediaID)
	if _, err := file.WriteString(data); err != nil {
		log.Printf("Помилка при запису в файл: %v\n", err)
	}

	log.Printf("Тригер збережено: %s\n", trigger)
	updateTriggerList()
}

func deleteTrigger(trigger string) {
	inputFile, err := os.Open("triggers.txt")
	if err != nil {
		log.Printf("Помилка при відкритті файлу: %v\n", err)
		return
	}
	defer inputFile.Close()

	lines := make([]string, 0)
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}
		if parts[0] != trigger {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Помилка при читанні файлу: %v\n", err)
		return
	}

	outputFile, err := os.Create("triggers.txt")
	if err != nil {
		log.Printf("Помилка при створенні файлу: %v\n", err)
		return
	}
	defer outputFile.Close()

	for _, line := range lines {
		_, err := outputFile.WriteString(line + "\n")
		if err != nil {
			log.Printf("Помилка при запису в файл: %v\n", err)
		}
	}
	log.Printf("Тригер видалено: %s\n", trigger)
	updateTriggerList()
}

func sendTriggeredMessage(bot *tgbotapi.BotAPI, chatID int64, text, triggerUser string) {
	file, err := os.Open("triggers.txt")
	if err != nil {
		log.Printf("Помилка при відкритті файлу: %v\n", err)
		return
	}
	defer file.Close()

	var bestMatchTrigger string
	var bestMatchMediaID string
	var bestMatchMediaType string
	var bestMatchMessageText string

	var matchingTriggers []string // Список підходящих триггерів

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "|", 4)
		if len(parts) < 4 {
			continue
		}
		trigger, messageText, mediaType, mediaID := parts[0], parts[1], parts[2], parts[3]

		// Перевіряємо, чи містить текст повідомлення цей триггер
		if strings.Contains(text, trigger) {
			matchingTriggers = append(matchingTriggers, trigger)

			// Якщо знайшли найкращий триггер, оновлюємо значення
			if len(trigger) > len(bestMatchTrigger) {
				bestMatchTrigger = trigger
				bestMatchMediaID = mediaID
				bestMatchMediaType = mediaType
				bestMatchMessageText = messageText
			}
		}
	}

	if len(matchingTriggers) > 0 {
		// Якщо є підходящі триггери, вибираємо найкращий
		// (можна додатково внести логіку, якщо є декілька підходящих)
		bestMatchTrigger = matchingTriggers[0]
	}

	if bestMatchTrigger != "" {
		// Якщо знайшли найкращий триггер, надсилаємо відповідне медіа
		sendMedia(bot, chatID, bestMatchMediaType, bestMatchMediaID, bestMatchMessageText, triggerUser)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Помилка при читанні файлу: %v\n", err)
	}
}

func sendMedia(bot *tgbotapi.BotAPI, chatID int64, mediaType, mediaID, caption, triggerUser string) {
	var msg tgbotapi.Chattable
	switch mediaType {
	case "photo":
		photo := tgbotapi.NewPhotoShare(chatID, mediaID)
		msg = photo
	case "sticker":
		msg = tgbotapi.NewStickerShare(chatID, mediaID)
	case "video":
		video := tgbotapi.NewVideoShare(chatID, mediaID)
		msg = video
	case "voice":
		voice := tgbotapi.NewVoiceShare(chatID, mediaID)
		msg = voice
	case "audio":
		audio := tgbotapi.NewAudioShare(chatID, mediaID)
		msg = audio
	case "animation":
		animation := tgbotapi.NewAnimationShare(chatID, mediaID)
		msg = animation
	case "videonote":
		msg = tgbotapi.NewVideoNoteShare(chatID, 0, mediaID) // Додано 0 як параметр для відеоповідомлення
	case "text":
		msg = tgbotapi.NewMessage(chatID, caption)
	default:
		log.Printf("Невідомий тип медіа: %v\n", mediaType)
		return
	}

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Помилка при відправці повідомлення: %v\n", err)
	}
}

func sendTriggerList(bot *tgbotapi.BotAPI, chatID int64) {
	updateTriggerList() // Оновлюємо список триггерів
	if len(triggers) == 0 {
		sendMessage(bot, chatID, "Список тригерів порожній.")
		return
	}

	triggerList := "Список тригерів:\n"
	for _, trigger := range triggers {
		triggerList += fmt.Sprintf(" - %s\n", trigger)
	}

	sendMessage(bot, chatID, triggerList)
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Помилка при відправці повідомлення: %v\n", err)
	}
}

func updateTriggerList() {
	fileData, err := os.ReadFile("triggers.txt")
	if err != nil {
		log.Printf("Помилка при читанні файлу: %v\n", err)
		return
	}

	triggerLines := strings.Split(string(fileData), "\n")
	triggers = make([]string, 0, len(triggerLines))

	for _, line := range triggerLines {
		// Ігноруємо порожні рядки
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) >= 1 {
			triggers = append(triggers, parts[0])
		}
	}
}

func isTriggerUnique(trigger string) bool {
	for _, t := range triggers {
		if t == trigger {
			return false
		}
	}
	return true
}

func isTriggerExists(trigger string) bool {
	for _, t := range triggers {
		if t == trigger {
			return true
		}
	}
	return false
}

func loadToken() (string, error) {
	// Якщо файл існує, читаємо токен з нього
	if _, err := os.Stat("config.txt"); err == nil {
		token, err := os.ReadFile("config.txt")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(token)), nil
	}

	// Якщо файл не існує, запитуємо користувача ввести токен
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Введіть API ключ бота: ")
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)

	// Зберігаємо токен у файл
	err := os.WriteFile("config.txt", []byte(token), 0644)
	if err != nil {
		return "", err
	}

	return token, nil
}
