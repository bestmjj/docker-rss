package main

import (
        "bytes"
        "encoding/json"
        "fmt"
        "log"
        "net/http"
        "os"
        "strings"
)

// Message 结构体，用于构建发送消息的请求体
type Message struct {
        ChatID    string `json:"chat_id"`
        Text      string `json:"text"`
        ParseMode string `json:"parse_mode"` // 支持 HTML 格式
}

func updateImages(updates []ImageUpdate) {
        // 从环境变量获取 Telegram Bot API Key 和 Chat ID
        apiKey := os.Getenv("TELEGRAM_API_KEY")
        chatID := os.Getenv("TELEGRAM_CHAT_ID")
        location := os.Getenv("LOCATION")

        // 检查环境变量是否已设置
        if apiKey == "" || chatID == "" {
                log.Fatal("环境变量 TELEGRAM_API_KEY 或 TELEGRAM_CHAT_ID 未设置")
        }

        // 构建 Telegram API URL
        apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", apiKey)

        // 构建消息内容
        var messageText strings.Builder
        messageText.WriteString(fmt.Sprintf("<b>📢 Image Update Notifications(%s)</b>\n\n", location))

        updateCount := 0
        for _, update := range updates {
                if update.UpdateAvailable {
                        updateCount++
                        log.Printf("update available for %s", update.ImageName)
                        //uniqueID := fmt.Sprintf("update-%s-%d-%d", update.ImageName, time.Now().UnixNano(), rand.Intn(10000))
                        // 添加到 Telegram 消息
                        messageText.WriteString(fmt.Sprintf("🔄 <b>Update for %s</b>\n", update.ImageName))
                        messageText.WriteString(fmt.Sprintf("  • <b>Current Hash:</b> %s\n", update.CurrentHash))
                        messageText.WriteString(fmt.Sprintf("  • <b>Latest Hash:</b> %s\n", update.LatestHash))
                        messageText.WriteString(fmt.Sprintf("  • <b>Architecture:</b> %s\n", update.Architecture))
                        messageText.WriteString(fmt.Sprintf("  • <b>Image Created:</b> %s\n\n", update.ImageCreated))
                }
        }

        // 如果没有更新，直接返回，不发送消息
        if updateCount == 0 {
                log.Println("No updates available, skipping Telegram message")
                return
        }

        // 准备要发送的消息，支持 HTML 格式
        message := Message{
                ChatID:    chatID,
                Text:      messageText.String(),
                ParseMode: "HTML", // 使用 HTML 格式以支持粗体等样式
        }

        // 将消息结构体转换为 JSON
        jsonData, err := json.Marshal(message)
        if err != nil {
                log.Printf("JSON 编码失败: %v", err)
                return
        }

        // 发送 HTTP POST 请求
        resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
                log.Printf("发送请求失败: %v", err)
                return
        }
        defer resp.Body.Close()

        // 检查响应状态码
        if resp.StatusCode != http.StatusOK {
                log.Printf("请求失败，状态码: %d", resp.StatusCode)
                return
        }

        // 读取响应内容（可选）
        var result map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
                log.Printf("解析响应失败: %v", err)
                return
        }

        // 打印成功信息
        log.Println("消息发送成功:", result)
}
