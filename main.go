package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var (
	replacer = strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
)

func main() {
	tgToken := flag.String("TELEGRAM_BOT_TOKEN", "", "Telegram Bot Token")
	apiKey := flag.String("GOOGLE_GEMINI_KEY", "", "API Key for genai")
	flag.Parse()

	if *tgToken == "" {
		*tgToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}

	if *tgToken == "" {
		log.Fatal("Telegram Bot Token must be set")
	}

	if *apiKey == "" {
		*apiKey = os.Getenv("GOOGLE_GEMINI_KEY")
	}

	if *apiKey == "" {
		log.Fatal("API Key must be set")
	}

	bot, err := tgbotapi.NewBotAPI(*tgToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(*apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Println("Error getting updates:", err)
	}

	for update := range updates {
		if update.Message != nil {
			if update.Message.Text != "" {
				go handleMessage(bot, client, update)
			} else if update.Message.Photo != nil {
				go handlePhoto(bot, update, client)
			}
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, client *genai.Client, update tgbotapi.Update) {
	if update.Message == nil {
		return // 确保消息存在
	}
	// 获取机器人的用户名
	botUsername, _ := bot.GetMe()

	// 构建可能的命令格式
	geminiCommand := "/gemini"
	geminiCommandWithBot := fmt.Sprintf("/gemini@%s", botUsername.UserName)
	// 检查是否是私聊或群组中的特定命令
	isPrivateChat := update.Message.Chat.IsPrivate()
	isGeminiCommand := strings.HasPrefix(update.Message.Text, geminiCommand) || strings.HasPrefix(update.Message.Text, geminiCommandWithBot)

	if isPrivateChat || isGeminiCommand {
		// 移除命令部分，保留查询内容
		userQuestion := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, geminiCommand))
		userQuestion = strings.TrimSpace(strings.TrimPrefix(userQuestion, geminiCommandWithBot))

		if userQuestion == "" && !isPrivateChat {
			replyWithResponse(bot, update, nil)
			return
		}

		ctx := context.Background()
		// 使用用户的提问作为输入生成内容
		resp, err := client.GenerativeModel("gemini-pro").GenerateContent(ctx, genai.Text(userQuestion))
		if err != nil {
			log.Println("Error generating content:", err)
			replyWithResponse(bot, update, nil)
			return
		}

		replyWithResponse(bot, update, resp)
	}
}

func handlePhoto(bot *tgbotapi.BotAPI, update tgbotapi.Update, client *genai.Client) {
	photos := *update.Message.Photo
	texts := update.Message.Text
	fmt.Println("texts from photo", texts)
	// 获取机器人的用户名
	botUsername, _ := bot.GetMe()

	// 构建可能的命令格式
	geminiCommandWithBot := fmt.Sprintf("/gemini@%s", botUsername.UserName)

	// 检查是否是私聊或群组中的特定命令
	isPrivateChat := update.Message.Chat.IsPrivate()
	isGeminiCommand := strings.HasPrefix(update.Message.Caption, "/gemini") || strings.HasPrefix(update.Message.Caption, geminiCommandWithBot)

	if isPrivateChat || isGeminiCommand {
		if len(photos) > 0 {
			photo := photos[len(photos)-1]

			// 获取图片的文件信息
			fileID := photo.FileID
			file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
			if err != nil {
				log.Println("Error getting file:", err)
				return
			}

			// 下载图片
			imgData, err := downloadImage(bot, file.FileID)
			if err != nil {
				log.Println("Error downloading image:", err)
				return
			}

			// 获取用户的文本输入
			userText := ""
			if update.Message.Caption != "" {
				userText = strings.TrimSpace(strings.TrimPrefix(update.Message.Caption, "/gemini"))
				userText = strings.TrimSpace(strings.TrimPrefix(userText, geminiCommandWithBot))
			}

			// 处理图片
			processAndReplyImage(bot, update, client, imgData, userText)
		}
	}
}

func processAndReplyImage(bot *tgbotapi.BotAPI, update tgbotapi.Update, client *genai.Client, imgData []byte, userText string) {
	ctx := context.Background()

	// 使用 Gemini API 处理图片
	model := client.GenerativeModel("gemini-pro-vision")
	prompt := []genai.Part{
		genai.ImageData("jpeg", imgData),
	}

	// 如果用户提供了文本描述，则添加到 prompt 中
	if userText != "" {
		prompt = append(prompt, genai.Text(userText))
	} else {
		// 如果没有用户输入的文本，可以使用默认文本或省略
		prompt = append(prompt, genai.Text("描述这张图片"))
	}

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		log.Println("Error generating content:", err)
		return
	}

	// 处理响应并回复用户
	replyWithResponse(bot, update, resp)
}

func replyWithResponse(bot *tgbotapi.BotAPI, update tgbotapi.Update, resp *genai.GenerateContentResponse) {
	// 假设我们只关心第一个响应
	if len(resp.Candidates) > 0 {
		messageText := printResponse(resp)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyToMessageID = update.Message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Println("Error sending message:", err)
		}
	}
}

func downloadImage(bot *tgbotapi.BotAPI, filePath string) ([]byte, error) {
	resp, err := bot.GetFileDirectURL(filePath)
	if err != nil {
		return nil, err
	}

	// 发起 HTTP GET 请求下载文件
	response, err := http.Get(resp)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// 读取响应内容
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func printResponse(resp *genai.GenerateContentResponse) string {
	var ret string
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			ret = ret + fmt.Sprintf("%v", part)
			fmt.Println(part)
		}
	}

	ret = replacer.Replace(ret)
	fmt.Println("---")
	return ret + "\n---"
}
