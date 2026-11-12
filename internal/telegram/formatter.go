package telegram

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

func escape(s string) string {
	return html.EscapeString(s)
}

func sanitizeTelegramButtonURL(u string) string {
	if u == "" {
		return ""
	}
	trimmed := strings.TrimSpace(u)
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return ""
	}
	// Telegram Bot API rejects literal word "localhost", but accepts "127.0.0.1", public domains, and public IPs.
	return strings.Replace(trimmed, "localhost", "127.0.0.1", -1)
}

func isValidPublicURL(u string) bool {
	if u == "" {
		return false
	}
	lower := strings.ToLower(u)
	if strings.Contains(lower, "localhost") || strings.Contains(lower, "127.0.0.1") || strings.Contains(lower, "0.0.0.0") {
		return false
	}
	if !strings.HasPrefix(lower, "https://") && !strings.HasPrefix(lower, "http://") {
		return false
	}
	hostPart := strings.TrimPrefix(strings.TrimPrefix(lower, "https://"), "http://")
	parts := strings.Split(hostPart, "/")
	return strings.Contains(parts[0], ".")
}

func formatPriorityLabel(p string) string {
	switch strings.ToLower(p) {
	case "critical":
		return "🔴 <b>P1 - CRITICAL</b> (&lt; 3 Jam)"
	case "high":
		return "🟠 <b>P2 - HIGH</b> (&lt; 8 Jam)"
	case "medium":
		return "🔵 <b>P3 - MEDIUM</b> (&lt; 24 Jam)"
	case "low":
		return "🟢 <b>P4 - LOW</b> (&lt; 48 Jam)"
	default:
		return escape(p)
	}
}

func formatCategoryLabel(c string) string {
	switch strings.ToLower(c) {
	case "sensor":
		return "🔧 Sensor & Flowmeter BBM"
	case "gateway":
		return "📡 Gateway IoT & Sinyal Satelit"
	case "voyage":
		return "🗺️ Rute & Jadwal Voyage"
	case "anomaly":
		return "⚠️ Anomali Bahan Bakar"
	case "account":
		return "🔐 Akun & Akses Pengguna"
	case "other":
		return "💬 Pertanyaan / Lainnya"
	default:
		return escape(c)
	}
}

func formatVesselStatus(status, port string) string {
	var s string
	switch strings.ToLower(status) {
	case "sailing":
		s = "🌊 Sedang Berlayar (At Sea)"
	case "docked":
		s = "⚓ Sedang Sandar (At Berth/Port)"
	case "anchorage":
		s = "⚓ Labuh Jangkar (At Anchorage)"
	default:
		s = "🚢 Operasional"
	}
	if port != "" {
		s += fmt.Sprintf(" - %s", escape(port))
	}
	return s
}

// FormatNewTicketMessage builds the incident card message for Telegram
func FormatNewTicketMessage(t *model.SupportTicket, dashboardURL string) (string, *InlineKeyboardMarkup) {
	companyName := "Global / Internal"
	if t.Company != nil && t.Company.Name != "" {
		companyName = t.Company.Name
	}

	vesselName := "Non-Spesifik (Umum)"
	if t.Vessel != nil && t.Vessel.Name != "" {
		vesselName = t.Vessel.Name
	}

	slaTargetStr := t.SlaTargetAt.In(time.Local).Format("02 Jan 15:04") + " WIB"

	var sb strings.Builder
	sb.WriteString("🚨 <b>[TIKET BARU MASUK] - #" + escape(t.TicketNumber) + "</b>\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("🏢 <b>Perusahaan:</b> " + escape(companyName) + "\n")
	sb.WriteString("👤 <b>Pelapor:</b> " + escape(t.ReporterName))
	if t.ReporterPhone != "" {
		sb.WriteString(" (" + escape(t.ReporterPhone) + ")")
	}
	sb.WriteString("\n")
	sb.WriteString("🚢 <b>Kapal:</b> " + escape(vesselName) + "\n")

	if t.VesselStatus != "" && t.VesselStatus != "unknown" {
		sb.WriteString("📍 <b>Status Kapal:</b> " + formatVesselStatus(t.VesselStatus, t.PortLocation) + "\n")
	}

	sb.WriteString("🏷️ <b>Kategori:</b> " + formatCategoryLabel(t.Category) + "\n")
	sb.WriteString("⚠️ <b>Urgensi:</b> " + formatPriorityLabel(t.Priority) + "\n")
	sb.WriteString("⏰ <b>Batas SLA:</b> " + escape(slaTargetStr) + "\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("📌 <b>Judul:</b>\n" + escape(t.Subject) + "\n\n")
	sb.WriteString("📝 <b>Kronologi / Uraian Masalah:</b>\n<i>" + escape(t.Description) + "</i>\n")

	if t.AttachmentURL != "" {
		sb.WriteString("\n📎 <b>Lampiran:</b> <i>File foto/berkas terlampir di bawah</i>\n")
	}

	sb.WriteString("──────────────────────────\n")
	sb.WriteString("Status: ⏳ <b>MENUNGGU RESPON TIM NOC</b>\n")
	sb.WriteString("<i>Silakan ambil tiket untuk konfirmasi penanganan langsung.</i>")

	row := []InlineKeyboardButton{
		{
			Text:         "🙋 Ambil Tiket (Accept)",
			CallbackData: fmt.Sprintf("claim:%d", t.ID),
		},
	}
	if btnURL := sanitizeTelegramButtonURL(dashboardURL); btnURL != "" {
		row = append(row, InlineKeyboardButton{
			Text: "🌐 Buka di Web",
			URL:  btnURL,
		})
	}

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{row},
	}

	return sb.String(), kb
}

// FormatClaimedTicketMessage updates the incident card when a technician accepts the ticket
func FormatClaimedTicketMessage(t *model.SupportTicket, claimedBy string, dashboardURL string) (string, *InlineKeyboardMarkup) {
	companyName := "Global / Internal"
	if t.Company != nil && t.Company.Name != "" {
		companyName = t.Company.Name
	}

	vesselName := "Non-Spesifik (Umum)"
	if t.Vessel != nil && t.Vessel.Name != "" {
		vesselName = t.Vessel.Name
	}

	ackTimeStr := time.Now().In(time.Local).Format("15:04") + " WIB"

	var sb strings.Builder
	sb.WriteString("⚙️ <b>[TIKET SEDANG DITANGANI] - #" + escape(t.TicketNumber) + "</b>\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("🏢 <b>Perusahaan:</b> " + escape(companyName) + "\n")
	sb.WriteString("👤 <b>Pelapor:</b> " + escape(t.ReporterName) + "\n")
	sb.WriteString("🚢 <b>Kapal:</b> " + escape(vesselName) + "\n")
	sb.WriteString("🏷️ <b>Kategori:</b> " + formatCategoryLabel(t.Category) + "\n")
	sb.WriteString("⚠️ <b>Urgensi:</b> " + formatPriorityLabel(t.Priority) + "\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("📌 <b>Judul:</b> " + escape(t.Subject) + "\n\n")
	sb.WriteString("📝 <b>Kronologi:</b>\n<i>" + escape(t.Description) + "</i>\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("Status: 🟡 <b>SEDANG DITANGANI (IN PROGRESS)</b>\n")
	sb.WriteString("👨‍💻 <b>Ditangani Oleh:</b> " + escape(claimedBy) + " (Pukul " + ackTimeStr + ")\n\n")
	sb.WriteString("💡 <i>Untuk membalas pelapor di web, langsung <b>Reply / Kutip</b> pesan ini.</i>")

	rowAction := []InlineKeyboardButton{
		{
			Text:         "✅ Selesaikan (Resolve)",
			CallbackData: fmt.Sprintf("resolve:%d", t.ID),
		},
	}
	if btnURL := sanitizeTelegramButtonURL(dashboardURL); btnURL != "" {
		rowAction = append(rowAction, InlineKeyboardButton{
			Text: "🌐 Buka di Web",
			URL:  btnURL,
		})
	}

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{
					Text:         "💬 Balas: Kutip Pesan Ini",
					CallbackData: "info_reply",
				},
			},
			rowAction,
		},
	}

	return sb.String(), kb
}

// FormatResolvedTicketMessage creates the resolution confirmation message
func FormatResolvedTicketMessage(t *model.SupportTicket, resolverName, notes, dashboardURL string) (string, *InlineKeyboardMarkup) {
	companyName := "Global / Internal"
	if t.Company != nil && t.Company.Name != "" {
		companyName = t.Company.Name
	}

	vesselName := "Non-Spesifik (Umum)"
	if t.Vessel != nil && t.Vessel.Name != "" {
		vesselName = t.Vessel.Name
	}

	resolveTimeStr := time.Now().In(time.Local).Format("02 Jan 15:04") + " WIB"

	var sb strings.Builder
	sb.WriteString("✅ <b>[TIKET SELESAI / RESOLVED] - #" + escape(t.TicketNumber) + "</b>\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("🏢 <b>Perusahaan:</b> " + escape(companyName) + "\n")
	sb.WriteString("🚢 <b>Kapal:</b> " + escape(vesselName) + "\n")
	sb.WriteString("👨‍💻 <b>Diselesaikan Oleh:</b> " + escape(resolverName) + "\n")
	sb.WriteString("⏱️ <b>Waktu Selesai:</b> " + resolveTimeStr + "\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("📌 <b>Judul:</b> " + escape(t.Subject) + "\n")
	sb.WriteString("📋 <b>Catatan Solusi:</b>\n<i>" + escape(notes) + "</i>\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("Status: 🟢 <b>TUNTAS (RESOLVED)</b>")

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{},
	}
	if btnURL := sanitizeTelegramButtonURL(dashboardURL); btnURL != "" {
		kb.InlineKeyboard = append(kb.InlineKeyboard, []InlineKeyboardButton{
			{
				Text: "🌐 Buka di Web",
				URL:  btnURL,
			},
		})
	}

	return sb.String(), kb
}

// FormatUserReplyNotification formats a notification when a client user posts a reply on Web
func FormatUserReplyNotification(t *model.SupportTicket, r *model.TicketReply) string {
	sender := r.SenderName
	if sender == "" {
		sender = "Pelapor"
	}
	if t.Company != nil && t.Company.Name != "" {
		sender += fmt.Sprintf(" (%s)", t.Company.Name)
	}

	timeStr := r.CreatedAt.In(time.Local).Format("15:04") + " WIB"

	var sb strings.Builder
	sb.WriteString("💬 <b>[TANGGAPAN BARU DARI KLIEN]</b>\n")
	sb.WriteString("No. Tiket: <b>#" + escape(t.TicketNumber) + "</b>\n")
	sb.WriteString("Dari: <b>" + escape(sender) + "</b> (" + timeStr + ")\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("<i>\"" + escape(r.Message) + "\"</i>\n")

	if r.AttachmentURL != "" {
		sb.WriteString("\n📎 <b>Lampiran:</b> " + escape(r.AttachmentURL) + "\n")
	}

	sb.WriteString("──────────────────────────\n")
	sb.WriteString("👉 <i>Reply pesan ini di Telegram untuk membalas ke klien.</i>")

	return sb.String()
}

// FormatMenuHub renders the central interactive control hub for NOC engineers
func FormatMenuHub(m *TicketMetricsSummary, callerName, webURL string) (string, *InlineKeyboardMarkup) {
	timeStr := time.Now().In(time.Local).Format("02 Jan 2006, 15:04") + " WIB"

	var sb strings.Builder
	sb.WriteString("🎛️ <b>[KIT NOC COMMAND DESK]</b>\n")
	sb.WriteString("<i>Sistem Pemantauan Operasional Armada & Telemetri</i>\n")
	sb.WriteString("──────────────────────────\n")
	if callerName != "" {
		sb.WriteString("👤 <b>Petugas:</b> " + escape(callerName) + "\n")
	}
	sb.WriteString("⏱️ <b>Waktu:</b> " + timeStr + "\n")
	sb.WriteString("⚡ <b>Status Gateway:</b> 🟢 <b>ONLINE SIAGA</b>\n")
	sb.WriteString("──────────────────────────\n")

	openCount := int64(0)
	critCount := int64(0)
	progCount := int64(0)
	if m != nil {
		openCount = m.OpenTickets
		critCount = m.CriticalP1
		progCount = m.InProgress
		slaOnTrack := float64(100)
		if m.TotalTickets > 0 && m.SlaBreached > 0 {
			slaOnTrack = float64(m.TotalTickets-m.SlaBreached) / float64(m.TotalTickets) * 100
		}
		sb.WriteString("📊 <b>Status Antrean Tiket:</b>\n")
		sb.WriteString(fmt.Sprintf("• ⏳ Menunggu Respon   : <b>%d Tiket</b>\n", openCount))
		sb.WriteString(fmt.Sprintf("• 🚨 Prioritas Kritis P1: <b>%d Tiket</b>\n", critCount))
		sb.WriteString(fmt.Sprintf("• ⚙️ Sedang Ditangani   : <b>%d Tiket</b>\n", progCount))
		sb.WriteString(fmt.Sprintf("• 🎯 Kepatuhan SLA       : <b>%.1f%%</b>\n", slaOnTrack))
	}
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("Silakan pilih menu navigasi di bawah:")

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{
					Text:         fmt.Sprintf("⏳ Tiket Menunggu (%d)", openCount),
					CallbackData: "menu_unhandled",
				},
				{
					Text:         fmt.Sprintf("🚨 Kritis P1 (%d)", critCount),
					CallbackData: "menu_critical",
				},
			},
			{
				{
					Text:         "📊 Ringkasan SLA",
					CallbackData: "menu_metrics",
				},
				{
					Text:         "📋 Semua Tiket Aktif",
					CallbackData: "menu_active",
				},
			},
		},
	}

	var row3 []InlineKeyboardButton
	if sanitized := sanitizeTelegramButtonURL(webURL); sanitized != "" {
		row3 = append(row3, InlineKeyboardButton{
			Text: "🌐 Buka di Web",
			URL:  sanitized,
		})
	}
	row3 = append(row3, InlineKeyboardButton{
		Text:         "🔄 Segarkan",
		CallbackData: "menu_main",
	})
	kb.InlineKeyboard = append(kb.InlineKeyboard, row3)

	return sb.String(), kb
}

// FormatUnhandledTicketsList renders the list of open tickets waiting for NOC response
func FormatUnhandledTicketsList(tickets []model.SupportTicket, webURL string) (string, *InlineKeyboardMarkup) {
	var sb strings.Builder
	sb.WriteString("⏳ <b>[TIKET MENUNGGU RESPON TIM NOC]</b>\n")
	sb.WriteString("──────────────────────────\n")

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{},
	}

	if len(tickets) == 0 {
		sb.WriteString("✅ <b>Luar biasa! Tidak ada tiket yang menunggu respon saat ini.</b>\n")
		sb.WriteString("Seluruh laporan kendala telah dikonfirmasi atau diselesaikan oleh tim NOC.\n")
	} else {
		sb.WriteString(fmt.Sprintf("Ditemukan <b>%d tiket</b> yang memerlukan konfirmasi segera:\n\n", len(tickets)))

		maxItems := 5
		if len(tickets) < maxItems {
			maxItems = len(tickets)
		}

		now := time.Now()
		for i := 0; i < maxItems; i++ {
			t := tickets[i]
			vessel := "Non-Spesifik"
			if t.Vessel != nil && t.Vessel.Name != "" {
				vessel = t.Vessel.Name
			}

			slaText := ""
			diff := t.SlaTargetAt.Sub(now)
			if diff > 0 {
				hours := int(diff.Hours())
				mins := int(diff.Minutes()) % 60
				slaText = fmt.Sprintf("⏱ Sisa SLA: %dj %dm", hours, mins)
			} else {
				slaText = "🚨 <b>SLA TERLEWATI</b>"
			}

			prioLabel := ""
			switch strings.ToLower(t.Priority) {
			case "critical":
				prioLabel = "🔴 P1"
			case "high":
				prioLabel = "🟠 P2"
			case "medium":
				prioLabel = "🔵 P3"
			default:
				prioLabel = "🟢 P4"
			}

			sb.WriteString(fmt.Sprintf("<b>%d. #%s</b> [%s]\n", i+1, escape(t.TicketNumber), prioLabel))
			sb.WriteString(fmt.Sprintf("   🚢 <b>Kapal:</b> %s\n", escape(vessel)))
			sb.WriteString(fmt.Sprintf("   📌 <b>Subjek:</b> %s\n", escape(t.Subject)))
			sb.WriteString(fmt.Sprintf("   %s\n\n", slaText))

			kb.InlineKeyboard = append(kb.InlineKeyboard, []InlineKeyboardButton{
				{
					Text:         fmt.Sprintf("🙋 Ambil #%s", t.TicketNumber),
					CallbackData: fmt.Sprintf("claim:%d", t.ID),
				},
			})
		}

		if len(tickets) > maxItems {
			sb.WriteString(fmt.Sprintf("<i>...dan %d tiket lainnya dapat dilihat di web dashboard.</i>\n", len(tickets)-maxItems))
		}
	}

	sb.WriteString("──────────────────────────\n")
	sb.WriteString("💡 <i>Klik tombol 'Ambil' di atas untuk mengonfirmasi tiket langsung.</i>")

	var navRow []InlineKeyboardButton
	navRow = append(navRow, InlineKeyboardButton{
		Text:         "🎛️ Menu Utama",
		CallbackData: "menu_main",
	})
	if sanitized := sanitizeTelegramButtonURL(webURL); sanitized != "" {
		navRow = append(navRow, InlineKeyboardButton{
			Text: "🌐 Buka Web",
			URL:  sanitized,
		})
	}
	navRow = append(navRow, InlineKeyboardButton{
		Text:         "🔄 Segarkan",
		CallbackData: "menu_unhandled",
	})
	kb.InlineKeyboard = append(kb.InlineKeyboard, navRow)

	return sb.String(), kb
}

// FormatCriticalTicketsList renders the list of active P1 tickets
func FormatCriticalTicketsList(tickets []model.SupportTicket, webURL string) (string, *InlineKeyboardMarkup) {
	var sb strings.Builder
	sb.WriteString("🚨 <b>[TIKET KRITIS PRIORITAS P1]</b>\n")
	sb.WriteString("<i>Target SLA: Wajib ditangani &lt; 3 Jam</i>\n")
	sb.WriteString("──────────────────────────\n")

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{},
	}

	if len(tickets) == 0 {
		sb.WriteString("🟢 <b>Aman terkendali.</b> Tidak ada tiket kritis P1 yang aktif saat ini.\n")
	} else {
		sb.WriteString(fmt.Sprintf("⚠️ <b>Perhatian! Ada %d tiket P1 aktif:</b>\n\n", len(tickets)))

		now := time.Now()
		for i, t := range tickets {
			vessel := "Non-Spesifik"
			if t.Vessel != nil && t.Vessel.Name != "" {
				vessel = t.Vessel.Name
			}

			statusStr := "⏳ Menunggu Respon"
			if t.Status == "in_progress" {
				handler := t.AcknowledgedBy
				if handler == "" {
					handler = "Teknisi"
				}
				statusStr = fmt.Sprintf("⚙️ Ditangani: %s", handler)
			}

			slaText := ""
			diff := t.SlaTargetAt.Sub(now)
			if diff > 0 {
				hours := int(diff.Hours())
				mins := int(diff.Minutes()) % 60
				slaText = fmt.Sprintf("⏱ Sisa SLA: %dj %dm", hours, mins)
			} else {
				slaText = "🚨 <b>SLA SUDAH TERLEWATI</b>"
			}

			sb.WriteString(fmt.Sprintf("<b>%d. #%s</b> (🚢 %s)\n", i+1, escape(t.TicketNumber), escape(vessel)))
			sb.WriteString(fmt.Sprintf("   📌 %s\n", escape(t.Subject)))
			sb.WriteString(fmt.Sprintf("   Status: %s\n", statusStr))
			sb.WriteString(fmt.Sprintf("   %s\n\n", slaText))

			if t.Status == "open" {
				kb.InlineKeyboard = append(kb.InlineKeyboard, []InlineKeyboardButton{
					{
						Text:         fmt.Sprintf("🙋 Ambil #%s", t.TicketNumber),
						CallbackData: fmt.Sprintf("claim:%d", t.ID),
					},
				})
			}
		}
	}

	sb.WriteString("──────────────────────────\n")

	var navRow []InlineKeyboardButton
	navRow = append(navRow, InlineKeyboardButton{
		Text:         "🎛️ Menu Utama",
		CallbackData: "menu_main",
	})
	if sanitized := sanitizeTelegramButtonURL(webURL); sanitized != "" {
		navRow = append(navRow, InlineKeyboardButton{
			Text: "🌐 Buka Web",
			URL:  sanitized,
		})
	}
	navRow = append(navRow, InlineKeyboardButton{
		Text:         "🔄 Segarkan",
		CallbackData: "menu_critical",
	})
	kb.InlineKeyboard = append(kb.InlineKeyboard, navRow)

	return sb.String(), kb
}

// FormatActiveTicketsList renders all tickets currently in progress or open
func FormatActiveTicketsList(tickets []model.SupportTicket, webURL string) (string, *InlineKeyboardMarkup) {
	var sb strings.Builder
	sb.WriteString("📋 <b>[DAFTAR SELURUH TIKET AKTIF]</b>\n")
	sb.WriteString("──────────────────────────\n")

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{},
	}

	if len(tickets) == 0 {
		sb.WriteString("🟢 Tidak ada tiket aktif saat ini. Seluruh tiket berstatus selesai / closed.\n")
	} else {
		maxItems := 6
		if len(tickets) < maxItems {
			maxItems = len(tickets)
		}

		for i := 0; i < maxItems; i++ {
			t := tickets[i]
			vessel := "Non-Spesifik"
			if t.Vessel != nil && t.Vessel.Name != "" {
				vessel = t.Vessel.Name
			}

			statusBadge := "⏳ Open"
			if t.Status == "in_progress" {
				statusBadge = "⚙️ Diproses"
			}

			sb.WriteString(fmt.Sprintf("<b>%d. #%s</b> [%s]\n", i+1, escape(t.TicketNumber), statusBadge))
			sb.WriteString(fmt.Sprintf("   🚢 %s • 📌 %s\n\n", escape(vessel), escape(t.Subject)))
		}

		if len(tickets) > maxItems {
			sb.WriteString(fmt.Sprintf("<i>...dan %d tiket aktif lainnya di web dashboard.</i>\n", len(tickets)-maxItems))
		}
	}

	sb.WriteString("──────────────────────────\n")

	var navRow []InlineKeyboardButton
	navRow = append(navRow, InlineKeyboardButton{
		Text:         "🎛️ Menu Utama",
		CallbackData: "menu_main",
	})
	if sanitized := sanitizeTelegramButtonURL(webURL); sanitized != "" {
		navRow = append(navRow, InlineKeyboardButton{
			Text: "🌐 Buka Web",
			URL:  sanitized,
		})
	}
	navRow = append(navRow, InlineKeyboardButton{
		Text:         "🔄 Segarkan",
		CallbackData: "menu_active",
	})
	kb.InlineKeyboard = append(kb.InlineKeyboard, navRow)

	return sb.String(), kb
}

// FormatMetricsSummary renders the SLA & Service Desk operational metrics card
func FormatMetricsSummary(m *TicketMetricsSummary, webURL string) (string, *InlineKeyboardMarkup) {
	timeStr := time.Now().In(time.Local).Format("02 Jan 2006, 15:04") + " WIB"

	var sb strings.Builder
	sb.WriteString("📊 <b>[STATUS LAYANAN & SLA NOC REALTIME]</b>\n")
	sb.WriteString("<i>Pembaruan Terkini: " + timeStr + "</i>\n")
	sb.WriteString("──────────────────────────\n")

	if m != nil {
		slaOnTrack := float64(100)
		if m.TotalTickets > 0 && m.SlaBreached > 0 {
			slaOnTrack = float64(m.TotalTickets-m.SlaBreached) / float64(m.TotalTickets) * 100
		}

		sb.WriteString("📋 <b>Volume Tiket:</b>\n")
		sb.WriteString(fmt.Sprintf("• Total Laporan Masuk    : <b>%d Tiket</b>\n", m.TotalTickets))
		sb.WriteString(fmt.Sprintf("• ⏳ Menunggu Respon     : <b>%d Tiket</b>\n", m.OpenTickets))
		sb.WriteString(fmt.Sprintf("• ⚙️ Sedang Ditangani    : <b>%d Tiket</b>\n", m.InProgress))
		sb.WriteString(fmt.Sprintf("• ✅ Selesai / Ditutup   : <b>%d Tiket</b>\n\n", m.Resolved))

		sb.WriteString("🎯 <b>Kepatuhan Target SLA:</b>\n")
		sb.WriteString(fmt.Sprintf("• 🚨 Melewati Batas SLA  : <b>%d Tiket</b>\n", m.SlaBreached))
		sb.WriteString(fmt.Sprintf("• 🟢 Kepatuhan SLA       : <b>%.1f%% On-Track</b>\n", slaOnTrack))
		sb.WriteString(fmt.Sprintf("• 🔴 Tiket Kritis P1     : <b>%d Tiket Aktif</b>\n", m.CriticalP1))
	} else {
		sb.WriteString("<i>Data metrik belum tersedia.</i>\n")
	}

	sb.WriteString("──────────────────────────\n")
	sb.WriteString("💡 <i>Target SLA EMERTI: P1 &lt; 3 Jam, P2 &lt; 8 Jam, P3 &lt; 24 Jam, P4 &lt; 48 Jam.</i>")

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{
					Text:         "🎛️ Menu Utama",
					CallbackData: "menu_main",
				},
				{
					Text:         "🔄 Segarkan Data",
					CallbackData: "menu_metrics",
				},
			},
		},
	}

	if sanitized := sanitizeTelegramButtonURL(webURL); sanitized != "" {
		kb.InlineKeyboard = append(kb.InlineKeyboard, []InlineKeyboardButton{
			{
				Text: "🌐 Lihat di Web Dashboard",
				URL:  sanitized,
			},
		})
	}

	return sb.String(), kb
}

// FormatHelpMessage renders the guide for bot usage
func FormatHelpMessage() (string, *InlineKeyboardMarkup) {
	var sb strings.Builder
	sb.WriteString("ℹ️ <b>[PANDUAN PENGGUNAAN BOT KIT NOC]</b>\n")
	sb.WriteString("<i>Pusat Operasional Maritim, Sensor BBM & Telemetri</i>\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString("Perintah cepat yang dapat diketik di grup:\n\n")
	sb.WriteString("• <code>/menu</code> — Buka Control Hub &amp; navigasi interaktif\n")
	sb.WriteString("• <code>/unhandled</code> — Tampilkan tiket yang menunggu konfirmasi\n")
	sb.WriteString("• <code>/kritis</code> — Tampilkan tiket darurat prioritas P1\n")
	sb.WriteString("• <code>/status</code> — Ringkasan metrik tiket &amp; target SLA\n")
	sb.WriteString("• <code>/help</code> — Bantuan penggunaan bot\n\n")
	sb.WriteString("💬 <b>Cara Membalas Klien Melalui Telegram:</b>\n")
	sb.WriteString("Cukup gunakan fitur <b>Reply / Kutip</b> pada pesan tiket kendala di grup ini. Balasan Anda akan otomatis diteruskan ke portal web klien secara realtime.\n")
	sb.WriteString("──────────────────────────")

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{
					Text:         "🎛️ Buka Menu Kontrol",
					CallbackData: "menu_main",
				},
			},
		},
	}

	return sb.String(), kb
}
