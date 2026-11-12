package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Client interface {
	SendMessage(chatID int64, text string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error)
	EditMessageText(chatID int64, messageID int, text string, replyMarkup *InlineKeyboardMarkup) error
	AnswerCallbackQuery(callbackQueryID string, text string, showAlert bool) error
	SendPhoto(chatID int64, photoURL string, caption string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error)
	SendPhotoFile(chatID int64, filePath string, caption string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error)
	SendDocument(chatID int64, docURL string, caption string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error)
	SendDocumentFile(chatID int64, filePath string, caption string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error)
	GetUpdates(offset int, timeout int) ([]Update, error)
	SetWebhook(url string) error
	DeleteWebhook() error
	GetMe() (*User, error)
	SetMyCommands(commands []BotCommand) error
}

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(token string) Client {
	return &client{
		token:   token,
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
		httpClient: &http.Client{
			Timeout: 40 * time.Second, // Long polling requires longer timeout than timeout param
		},
	}
}

func (c *client) SendMessage(chatID int64, text string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
		"disable_web_page_preview": true,
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	if replyToMsgID != nil && *replyToMsgID > 0 {
		payload["reply_to_message_id"] = *replyToMsgID
	}

	return c.postAndDecodeMessage("sendMessage", payload)
}

func (c *client) EditMessageText(chatID int64, messageID int, text string, replyMarkup *InlineKeyboardMarkup) error {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
		"disable_web_page_preview": true,
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}

	_, err := c.postAndDecodeMessage("editMessageText", payload)
	return err
}

func (c *client) AnswerCallbackQuery(callbackQueryID string, text string, showAlert bool) error {
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
		"show_alert":        showAlert,
	}
	var res ApiResponse[bool]
	return c.post("answerCallbackQuery", payload, &res)
}

func (c *client) SendPhoto(chatID int64, photoURL string, caption string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"photo":      photoURL,
		"caption":    caption,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	if replyToMsgID != nil && *replyToMsgID > 0 {
		payload["reply_to_message_id"] = *replyToMsgID
	}

	return c.postAndDecodeMessage("sendPhoto", payload)
}

func (c *client) SendDocument(chatID int64, docURL string, caption string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"document":   docURL,
		"caption":    caption,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	if replyToMsgID != nil && *replyToMsgID > 0 {
		payload["reply_to_message_id"] = *replyToMsgID
	}

	return c.postAndDecodeMessage("sendDocument", payload)
}

func (c *client) SendPhotoFile(chatID int64, filePath string, caption string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error) {
	return c.sendLocalMedia("sendPhoto", "photo", chatID, filePath, caption, replyMarkup, replyToMsgID)
}

func (c *client) SendDocumentFile(chatID int64, filePath string, caption string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error) {
	return c.sendLocalMedia("sendDocument", "document", chatID, filePath, caption, replyMarkup, replyToMsgID)
}

func (c *client) sendLocalMedia(method, fieldName string, chatID int64, filePath string, caption string, replyMarkup *InlineKeyboardMarkup, replyToMsgID *int) (*Message, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local media file: %w", err)
	}
	defer file.Close()

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	if err := w.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
		return nil, err
	}
	if caption != "" {
		_ = w.WriteField("caption", caption)
		_ = w.WriteField("parse_mode", "HTML")
	}
	if replyToMsgID != nil && *replyToMsgID > 0 {
		_ = w.WriteField("reply_to_message_id", strconv.Itoa(*replyToMsgID))
	}
	if replyMarkup != nil {
		if rmBytes, err := json.Marshal(replyMarkup); err == nil {
			_ = w.WriteField("reply_markup", string(rmBytes))
		}
	}

	part, err := w.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", c.baseURL, method)
	req, err := http.NewRequest(http.MethodPost, url, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res ApiResponse[Message]
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", string(respBytes))
	}
	if !res.Ok {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", res.Description, res.ErrorCode)
	}
	return &res.Result, nil
}

func (c *client) GetUpdates(offset int, timeout int) ([]Update, error) {
	payload := map[string]interface{}{
		"offset":  offset,
		"timeout": timeout,
		"allowed_updates": []string{"message", "callback_query"},
	}

	var res ApiResponse[[]Update]
	if err := c.post("getUpdates", payload, &res); err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", res.Description, res.ErrorCode)
	}
	return res.Result, nil
}

func (c *client) SetWebhook(url string) error {
	payload := map[string]interface{}{
		"url":             url,
		"allowed_updates": []string{"message", "callback_query"},
	}
	var res ApiResponse[bool]
	if err := c.post("setWebhook", payload, &res); err != nil {
		return err
	}
	if !res.Ok {
		return fmt.Errorf("telegram API error: %s", res.Description)
	}
	return nil
}

func (c *client) DeleteWebhook() error {
	payload := map[string]interface{}{
		"drop_pending_updates": false,
	}
	var res ApiResponse[bool]
	if err := c.post("deleteWebhook", payload, &res); err != nil {
		return err
	}
	if !res.Ok {
		return fmt.Errorf("telegram API error: %s", res.Description)
	}
	return nil
}

func (c *client) GetMe() (*User, error) {
	var res ApiResponse[User]
	if err := c.post("getMe", map[string]interface{}{}, &res); err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, fmt.Errorf("telegram API error: %s", res.Description)
	}
	return &res.Result, nil
}

func (c *client) postAndDecodeMessage(method string, payload interface{}) (*Message, error) {
	var res ApiResponse[Message]
	if err := c.post(method, payload, &res); err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", res.Description, res.ErrorCode)
	}
	return &res.Result, nil
}

func (c *client) post(method string, payload interface{}, out interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s", c.baseURL, method)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(respBytes, out); err != nil {
		return errors.New("failed to parse response from Telegram: " + string(respBytes))
	}

	return nil
}

func (c *client) SetMyCommands(commands []BotCommand) error {
	payload := map[string]interface{}{
		"commands": commands,
	}
	var res ApiResponse[bool]
	if err := c.post("setMyCommands", payload, &res); err != nil {
		return err
	}
	if !res.Ok {
		return fmt.Errorf("telegram API error: %s", res.Description)
	}
	return nil
}
