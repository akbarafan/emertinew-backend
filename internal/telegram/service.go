package telegram

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

type TicketMetricsSummary struct {
	TotalTickets int64
	OpenTickets  int64
	CriticalP1   int64
	InProgress   int64
	Resolved     int64
	SlaBreached  int64
}

type TicketActionHandler interface {
	GetTicketByIDModel(id uint) (*model.SupportTicket, error)
	GetTicketByTelegramMessageID(msgID int) (*model.SupportTicket, error)
	AcknowledgeTicketFromTelegram(ticketID uint, adminName string) (*model.SupportTicket, error)
	ReplyTicketFromTelegram(ticketID uint, replyText string, senderName string, tgMsgID int) (*model.TicketReply, error)
	ResolveTicketFromTelegram(ticketID uint, adminName string, notes string) (*model.SupportTicket, error)
	GetUnhandledTickets() ([]model.SupportTicket, error)
	GetCriticalTickets() ([]model.SupportTicket, error)
	GetActiveTickets() ([]model.SupportTicket, error)
	GetNOCMetrics() (*TicketMetricsSummary, error)
}

type Service interface {
	IsEnabled() bool
	SetTicketHandler(handler TicketActionHandler)
	NotifyNewTicket(ticket *model.SupportTicket) (int, error)
	NotifyTicketClaimed(ticket *model.SupportTicket, staffName string) error
	NotifyTicketReply(ticket *model.SupportTicket, reply *model.TicketReply) (int, error)
	NotifyTicketResolved(ticket *model.SupportTicket, resolverName, notes string) error
	HandleUpdate(update Update) error
	StartPolling()
	StopPolling()
}

type service struct {
	client        Client
	token         string
	chatID        int64
	webhookURL    string
	appBaseURL    string
	ticketHandler TicketActionHandler

	// Polling control
	stopChan   chan struct{}
	pollingMux sync.Mutex
	isPolling  bool
}

func NewService(token, chatIDStr, webhookURL, appBaseURL string) Service {
	var chatID int64
	if chatIDStr != "" {
		if id, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
			chatID = id
		} else {
			log.Printf("[TELEGRAM] Warning: Invalid TELEGRAM_SUPPORT_CHAT_ID: %v", err)
		}
	}

	var cl Client
	if token != "" {
		cl = NewClient(token)
	}

	s := &service{
		client:     cl,
		token:      token,
		chatID:     chatID,
		webhookURL: webhookURL,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
		stopChan:   make(chan struct{}),
	}

	return s
}

func (s *service) IsEnabled() bool {
	return s.client != nil && s.token != "" && s.chatID != 0
}

func (s *service) SetTicketHandler(handler TicketActionHandler) {
	s.ticketHandler = handler
}

func (s *service) buildDashboardURL(ticketID uint) string {
	if s.appBaseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/support/admin", s.appBaseURL)
}

func (s *service) NotifyNewTicket(t *model.SupportTicket) (int, error) {
	if !s.IsEnabled() {
		return 0, nil
	}

	dashboardURL := s.buildDashboardURL(t.ID)
	text, markup := FormatNewTicketMessage(t, dashboardURL)

	msg, err := s.client.SendMessage(s.chatID, text, markup, nil)
	if err != nil {
		log.Printf("[TELEGRAM] Failed to send new ticket alert: %v", err)
		return 0, err
	}

	// Dispatch local photo/document attachment directly to the Telegram chat
	if t.AttachmentURL != "" {
		go s.dispatchLocalAttachment(t.AttachmentURL, t.TicketNumber, msg.MessageID)
	}

	return msg.MessageID, nil
}

func (s *service) dispatchLocalAttachment(attachmentURL, ticketNumber string, replyToMsgID int) {
	if !s.IsEnabled() || attachmentURL == "" || replyToMsgID <= 0 {
		return
	}

	cleanRel := strings.TrimPrefix(attachmentURL, "/")
	baseName := filepath.Base(attachmentURL)
	possiblePaths := []string{
		filepath.Join("assets", cleanRel),
		filepath.Join(".", "assets", cleanRel),
		filepath.Join("backend", "assets", cleanRel),
		filepath.Join("assets", "uploads", "tickets", baseName),
		filepath.Join("backend", "assets", "uploads", "tickets", baseName),
		filepath.Join(".", cleanRel),
	}

	var foundPath string
	for _, p := range possiblePaths {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			foundPath = p
			break
		}
	}

	if foundPath == "" {
		log.Printf("[TELEGRAM] Attachment file not found on disk for ticket %s: %s", ticketNumber, attachmentURL)
		return
	}

	ext := strings.ToLower(filepath.Ext(foundPath))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		caption := fmt.Sprintf("📷 <b>Lampiran Tiket #%s</b>\n<i>Nama berkas: %s</i>", escape(ticketNumber), escape(filepath.Base(foundPath)))
		_, err := s.client.SendPhotoFile(s.chatID, foundPath, caption, nil, &replyToMsgID)
		if err != nil {
			log.Printf("[TELEGRAM] Failed to send photo attachment: %v", err)
		} else {
			log.Printf("[TELEGRAM] ✅ Photo attachment dispatched to Telegram for ticket #%s", ticketNumber)
		}
	default:
		caption := fmt.Sprintf("📄 <b>Lampiran Berkas Tiket #%s</b>\n<i>Nama berkas: %s</i>", escape(ticketNumber), escape(filepath.Base(foundPath)))
		_, err := s.client.SendDocumentFile(s.chatID, foundPath, caption, nil, &replyToMsgID)
		if err != nil {
			log.Printf("[TELEGRAM] Failed to send document attachment: %v", err)
		} else {
			log.Printf("[TELEGRAM] ✅ Document attachment dispatched to Telegram for ticket #%s", ticketNumber)
		}
	}
}

func (s *service) NotifyTicketClaimed(t *model.SupportTicket, staffName string) error {
	if !s.IsEnabled() || t.TelegramMessageID == nil {
		return nil
	}

	dashboardURL := s.buildDashboardURL(t.ID)
	text, markup := FormatClaimedTicketMessage(t, staffName, dashboardURL)

	err := s.client.EditMessageText(s.chatID, *t.TelegramMessageID, text, markup)
	if err != nil {
		log.Printf("[TELEGRAM] Failed to edit claimed ticket message: %v", err)
		return err
	}
	return nil
}

func (s *service) NotifyTicketReply(t *model.SupportTicket, r *model.TicketReply) (int, error) {
	if !s.IsEnabled() || t.TelegramMessageID == nil {
		return 0, nil
	}

	text := FormatUserReplyNotification(t, r)
	msg, err := s.client.SendMessage(s.chatID, text, nil, t.TelegramMessageID)
	if err != nil {
		log.Printf("[TELEGRAM] Failed to send reply notification: %v", err)
		return 0, err
	}

	// Dispatch local photo/document attachment in reply if attached
	if r.AttachmentURL != "" {
		go s.dispatchLocalAttachment(r.AttachmentURL, t.TicketNumber, msg.MessageID)
	}

	return msg.MessageID, nil
}

func (s *service) NotifyTicketResolved(t *model.SupportTicket, resolverName, notes string) error {
	if !s.IsEnabled() || t.TelegramMessageID == nil {
		return nil
	}

	dashboardURL := s.buildDashboardURL(t.ID)
	text, markup := FormatResolvedTicketMessage(t, resolverName, notes, dashboardURL)

	// Send as threaded resolution message
	_, err := s.client.SendMessage(s.chatID, text, markup, t.TelegramMessageID)
	if err != nil {
		log.Printf("[TELEGRAM] Failed to send resolved notification: %v", err)
		return err
	}
	return nil
}

func (s *service) HandleUpdate(update Update) error {
	if s.ticketHandler == nil {
		return errors.New("ticket handler not set")
	}

	// 1. Handle Callback Queries (Button Clicks)
	if update.CallbackQuery != nil {
		return s.handleCallbackQuery(update.CallbackQuery)
	}

	// 2. Handle Messages (Replies from staff)
	if update.Message != nil {
		return s.handleMessage(update.Message)
	}

	return nil
}

func (s *service) handleCallbackQuery(cb *CallbackQuery) error {
	data := cb.Data
	staffName := cb.From.DisplayName()

	if strings.HasPrefix(data, "claim:") {
		idStr := strings.TrimPrefix(data, "claim:")
		idUint, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			_ = s.client.AnswerCallbackQuery(cb.ID, "ID Tiket tidak valid", true)
			return err
		}

		ticket, err := s.ticketHandler.GetTicketByIDModel(uint(idUint))
		if err != nil || ticket == nil {
			_ = s.client.AnswerCallbackQuery(cb.ID, "Tiket tidak ditemukan di sistem.", true)
			return err
		}

		if ticket.Status != "open" {
			handledBy := ticket.AcknowledgedBy
			if handledBy == "" {
				handledBy = "teknisi lain"
			}
			_ = s.client.AnswerCallbackQuery(cb.ID, fmt.Sprintf("Tiket ini sudah ditangani oleh %s!", handledBy), true)
			return nil
		}

		// Update ticket in database
		updated, err := s.ticketHandler.AcknowledgeTicketFromTelegram(uint(idUint), staffName)
		if err != nil {
			_ = s.client.AnswerCallbackQuery(cb.ID, "Gagal konfirmasi penanganan: "+err.Error(), true)
			return err
		}

		// Acknowledge to user
		_ = s.client.AnswerCallbackQuery(cb.ID, fmt.Sprintf("✅ Anda resmi menangani tiket #%s!", updated.TicketNumber), false)

		// Edit original Telegram message
		if cb.Message != nil {
			dashboardURL := s.buildDashboardURL(updated.ID)
			text, markup := FormatClaimedTicketMessage(updated, staffName, dashboardURL)
			_ = s.client.EditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, markup)
		}
		return nil
	}

	if strings.HasPrefix(data, "resolve:") {
		idStr := strings.TrimPrefix(data, "resolve:")
		idUint, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			_ = s.client.AnswerCallbackQuery(cb.ID, "ID Tiket tidak valid", true)
			return err
		}

		ticket, err := s.ticketHandler.GetTicketByIDModel(uint(idUint))
		if err != nil || ticket == nil {
			_ = s.client.AnswerCallbackQuery(cb.ID, "Tiket tidak ditemukan.", true)
			return err
		}

		if ticket.Status == "resolved" || ticket.Status == "closed" {
			_ = s.client.AnswerCallbackQuery(cb.ID, "Tiket ini sudah tuntas / ditutup.", true)
			return nil
		}

		notes := fmt.Sprintf("Diselesaikan oleh tim NOC (%s)", staffName)
		updated, err := s.ticketHandler.ResolveTicketFromTelegram(uint(idUint), staffName, notes)
		if err != nil {
			_ = s.client.AnswerCallbackQuery(cb.ID, "Gagal menyelesaikan tiket: "+err.Error(), true)
			return err
		}

		_ = s.client.AnswerCallbackQuery(cb.ID, fmt.Sprintf("✅ Tiket #%s berhasil diselesaikan!", updated.TicketNumber), false)

		dashboardURL := s.buildDashboardURL(updated.ID)
		text, markup := FormatResolvedTicketMessage(updated, staffName, notes, dashboardURL)
		if cb.Message != nil {
			_, _ = s.client.SendMessage(cb.Message.Chat.ID, text, markup, &cb.Message.MessageID)
		}
		return nil
	}

	if data == "info_reply" {
		_ = s.client.AnswerCallbackQuery(cb.ID, "Untuk membalas pelapor di portal web, gunakan fitur Reply/Kutip pada pesan tiket ini.", true)
		return nil
	}

	// 3. Menu Navigation Callbacks
	dashboardURL := s.buildDashboardURL(0)

	switch data {
	case "menu_main":
		metrics, _ := s.ticketHandler.GetNOCMetrics()
		text, markup := FormatMenuHub(metrics, staffName, dashboardURL)
		_ = s.client.AnswerCallbackQuery(cb.ID, "Memuat Menu Hub NOC...", false)
		if cb.Message != nil {
			return s.client.EditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, markup)
		}
		return nil

	case "menu_unhandled":
		tickets, _ := s.ticketHandler.GetUnhandledTickets()
		text, markup := FormatUnhandledTicketsList(tickets, dashboardURL)
		_ = s.client.AnswerCallbackQuery(cb.ID, fmt.Sprintf("%d tiket menunggu respon", len(tickets)), false)
		if cb.Message != nil {
			return s.client.EditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, markup)
		}
		return nil

	case "menu_critical":
		tickets, _ := s.ticketHandler.GetCriticalTickets()
		text, markup := FormatCriticalTicketsList(tickets, dashboardURL)
		_ = s.client.AnswerCallbackQuery(cb.ID, fmt.Sprintf("%d tiket prioritas P1 aktif", len(tickets)), false)
		if cb.Message != nil {
			return s.client.EditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, markup)
		}
		return nil

	case "menu_active":
		tickets, _ := s.ticketHandler.GetActiveTickets()
		text, markup := FormatActiveTicketsList(tickets, dashboardURL)
		_ = s.client.AnswerCallbackQuery(cb.ID, fmt.Sprintf("%d tiket aktif", len(tickets)), false)
		if cb.Message != nil {
			return s.client.EditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, markup)
		}
		return nil

	case "menu_metrics":
		metrics, _ := s.ticketHandler.GetNOCMetrics()
		text, markup := FormatMetricsSummary(metrics, dashboardURL)
		_ = s.client.AnswerCallbackQuery(cb.ID, "Memuat metrik SLA terkini...", false)
		if cb.Message != nil {
			return s.client.EditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, markup)
		}
		return nil

	case "menu_help":
		text, markup := FormatHelpMessage()
		_ = s.client.AnswerCallbackQuery(cb.ID, "Panduan penggunaan bot", false)
		if cb.Message != nil {
			return s.client.EditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, markup)
		}
		return nil
	}

	return nil
}

func (s *service) handleMessage(msg *Message) error {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return nil
	}

	// 1. Handle Slash Commands (/menu, /unhandled, /status, /kritis, /help, /start)
	if strings.HasPrefix(text, "/") {
		return s.handleCommand(msg, text)
	}

	// 2. Only handle replies that quote a ticket message
	if msg.ReplyToMessage == nil {
		return nil
	}

	parentMsgID := msg.ReplyToMessage.MessageID
	ticket, err := s.ticketHandler.GetTicketByTelegramMessageID(parentMsgID)
	if err != nil || ticket == nil {
		// Not a message quoting a ticket
		return nil
	}

	if ticket.Status == "closed" {
		replyText := "⚠️ Tiket #" + ticket.TicketNumber + " sudah ditutup resmi, balasan tidak dapat diteruskan."
		_, _ = s.client.SendMessage(msg.Chat.ID, replyText, nil, &msg.MessageID)
		return nil
	}

	staffName := msg.From.DisplayName()
	_, err = s.ticketHandler.ReplyTicketFromTelegram(ticket.ID, msg.Text, staffName, msg.MessageID)
	if err != nil {
		log.Printf("[TELEGRAM] Failed to record reply from Telegram: %v", err)
		return err
	}

	// Confirm to Telegram thread
	confirmText := fmt.Sprintf("✅ <i>Balasan dari %s berhasil diteruskan ke portal web klien.</i>", escape(staffName))
	_, _ = s.client.SendMessage(msg.Chat.ID, confirmText, nil, &msg.MessageID)

	return nil
}

func (s *service) handleCommand(msg *Message, text string) error {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return nil
	}
	cmd := strings.ToLower(fields[0])
	cmd = strings.TrimPrefix(cmd, "/")
	cmd = strings.Split(cmd, "@")[0]

	dashboardURL := s.buildDashboardURL(0)
	senderName := msg.From.DisplayName()

	switch cmd {
	case "menu", "start":
		metrics, _ := s.ticketHandler.GetNOCMetrics()
		replyText, markup := FormatMenuHub(metrics, senderName, dashboardURL)
		_, err := s.client.SendMessage(msg.Chat.ID, replyText, markup, &msg.MessageID)
		return err

	case "unhandled", "open", "antrean":
		tickets, _ := s.ticketHandler.GetUnhandledTickets()
		replyText, markup := FormatUnhandledTicketsList(tickets, dashboardURL)
		_, err := s.client.SendMessage(msg.Chat.ID, replyText, markup, &msg.MessageID)
		return err

	case "kritis", "critical", "urgent", "p1":
		tickets, _ := s.ticketHandler.GetCriticalTickets()
		replyText, markup := FormatCriticalTicketsList(tickets, dashboardURL)
		_, err := s.client.SendMessage(msg.Chat.ID, replyText, markup, &msg.MessageID)
		return err

	case "active", "aktif", "semua":
		tickets, _ := s.ticketHandler.GetActiveTickets()
		replyText, markup := FormatActiveTicketsList(tickets, dashboardURL)
		_, err := s.client.SendMessage(msg.Chat.ID, replyText, markup, &msg.MessageID)
		return err

	case "status", "metrics", "ringkasan":
		metrics, _ := s.ticketHandler.GetNOCMetrics()
		replyText, markup := FormatMetricsSummary(metrics, dashboardURL)
		_, err := s.client.SendMessage(msg.Chat.ID, replyText, markup, &msg.MessageID)
		return err

	case "help", "panduan", "bantuan":
		replyText, markup := FormatHelpMessage()
		_, err := s.client.SendMessage(msg.Chat.ID, replyText, markup, &msg.MessageID)
		return err
	}

	return nil
}

func (s *service) StartPolling() {
	s.pollingMux.Lock()
	if s.isPolling || !s.IsEnabled() {
		s.pollingMux.Unlock()
		return
	}
	s.isPolling = true
	s.pollingMux.Unlock()

	// If webhook URL is set, register webhook instead of long polling
	if s.webhookURL != "" {
		log.Printf("[TELEGRAM] Setting webhook to %s", s.webhookURL)
		if err := s.client.SetWebhook(s.webhookURL); err != nil {
			log.Printf("[TELEGRAM] Warning: Failed to set webhook: %v", err)
		} else {
			log.Println("[TELEGRAM] Webhook successfully registered")
			return
		}
	} else {
		// Remove existing webhook before long polling
		_ = s.client.DeleteWebhook()
	}

	log.Println("[TELEGRAM] Starting long polling worker...")
	go func() {
		// Register slash commands so the Menu button appears in Telegram clients
		_ = s.client.SetMyCommands([]BotCommand{
			{Command: "menu", Description: "🎛️ Buka Control Hub & Menu NOC"},
			{Command: "unhandled", Description: "⏳ Tiket menunggu respon NOC"},
			{Command: "kritis", Description: "🚨 Tiket kritis prioritas P1"},
			{Command: "status", Description: "📊 Ringkasan SLA & metrik"},
			{Command: "help", Description: "ℹ️ Panduan penggunaan bot"},
		})

		offset := 0
		for {
			select {
			case <-s.stopChan:
				log.Println("[TELEGRAM] Polling worker stopped")
				return
			default:
				updates, err := s.client.GetUpdates(offset, 20)
				if err != nil {
					// Exponential delay on network issues
					time.Sleep(5 * time.Second)
					continue
				}

				for _, u := range updates {
					if u.UpdateID >= offset {
						offset = u.UpdateID + 1
					}
					if err := s.HandleUpdate(u); err != nil {
						log.Printf("[TELEGRAM] Error handling update %d: %v", u.UpdateID, err)
					}
				}
			}
		}
	}()
}

func (s *service) StopPolling() {
	s.pollingMux.Lock()
	defer s.pollingMux.Unlock()
	if s.isPolling {
		close(s.stopChan)
		s.isPolling = false
	}
}
