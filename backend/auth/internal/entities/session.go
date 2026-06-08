package entities

import "time"

// SessionInfo хранит данные об устройстве и сети, с которого была создана сессия.
// Иммутабельные поля (IP, геолокация, устройство, CreatedAt) записываются при
// создании токена и больше не изменяются.
// Изменяемые поля (LastIP, LastActiveAt) обновляются при каждом RefreshToken.
type SessionInfo struct {
	// ── Сеть ──────────────────────────────────────────────────────────────────
	IP     string // IP при создании сессии (иммутабельно)
	LastIP string // IP последнего RefreshToken
	ISP    string // Интернет-провайдер / название ASN

	// ── Геолокация (по IP, из MaxMind GeoLite2) ───────────────────────────────
	CountryCode string // ISO 3166-1 alpha-2: "RU", "US", "DE"
	CountryName string // "Russia", "United States"
	City        string // "Moscow", "Berlin"
	Timezone    string // "Europe/Moscow", "America/New_York"

	// ── Устройство (из User-Agent) ────────────────────────────────────────────
	DeviceType     string // "desktop" | "mobile" | "tablet"
	OS             string // "Windows", "macOS", "iOS", "Android", "Linux"
	OSVersion      string // "11", "14.5", "17.4"
	Browser        string // "Chrome", "Safari", "Firefox", "Edge"
	BrowserVersion string // "125.0"
	UserAgentRaw   string // полная строка User-Agent (запасной вариант)

	// ── Временны́е метки ──────────────────────────────────────────────────────
	CreatedAt    time.Time // Когда создана сессия (иммутабельно)
	LastActiveAt time.Time // Последний вызов RefreshToken (обновляется)
}
