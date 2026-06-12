package entities

import (
	"strconv"
	"time"

	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
)

// SessionInfo хранит данные об устройстве и сети, с которого была создана сессия.
// Иммутабельные поля (IP, геолокация, устройство, CreatedAt) записываются при
// создании токена и больше не изменяются.
// Изменяемые поля (LastIP, LastActiveAt) обновляются при каждом RefreshToken.
type SessionInfo struct {
	// ── Сеть ──────────────────────────────────────────────────────────────────
	IP     string // IP при создании сессии (иммутабельно)
	LastIP string // IP последнего RefreshToken
	ISP    string // Интернет-провайдер / название ASN

	// ── Геолокация (по IP) ────────────────────────────────────────────────────
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

	// ── Временны́е метки ───────────────────────────────────────────────────────
	CreatedAt    time.Time // Когда создана сессия (иммутабельно)
	LastActiveAt time.Time // Последний вызов RefreshToken (обновляется)
}

// ── Конверторы ────────────────────────────────────────────────────────────────

// FromProto конвертирует proto SessionInfo в доменную структуру.
// Если session == nil (клиент не передал данные), возвращает пустую структуру
// с CreatedAt/LastActiveAt = time.Now(), чтобы Redis хранил хотя бы метки времени.
func (e *SessionInfo) FromProto(s *pb.SessionInfo) {
	now := time.Now()
	if s == nil {
		e.CreatedAt = now
		e.LastActiveAt = now
		return
	}
	createdAt := now
	if s.GetCreatedAt() != 0 {
		createdAt = time.Unix(s.GetCreatedAt(), 0)
	}
	lastActiveAt := now
	if s.GetLastActiveAt() != 0 {
		lastActiveAt = time.Unix(s.GetLastActiveAt(), 0)
	}

	e.IP = s.GetIp()
	e.LastIP = s.GetLastIp()
	e.ISP = s.GetIsp()
	e.CountryCode = s.GetCountryCode()
	e.CountryName = s.GetCountryName()
	e.City = s.GetCity()
	e.Timezone = s.GetTimezone()
	e.DeviceType = s.GetDeviceType()
	e.OS = s.GetOs()
	e.OSVersion = s.GetOsVersion()
	e.Browser = s.GetBrowser()
	e.BrowserVersion = s.GetBrowserVersion()
	e.UserAgentRaw = s.GetUserAgentRaw()
	e.CreatedAt = createdAt
	e.LastActiveAt = lastActiveAt
}

// ToProto конвертирует доменную структуру SessionInfo в proto-сообщение.
func (e *SessionInfo) ToProto() *pb.SessionInfo {
	return &pb.SessionInfo{
		Ip:             e.IP,
		LastIp:         e.LastIP,
		Isp:            e.ISP,
		CountryCode:    e.CountryCode,
		CountryName:    e.CountryName,
		City:           e.City,
		Timezone:       e.Timezone,
		DeviceType:     e.DeviceType,
		Os:             e.OS,
		OsVersion:      e.OSVersion,
		Browser:        e.Browser,
		BrowserVersion: e.BrowserVersion,
		UserAgentRaw:   e.UserAgentRaw,
		CreatedAt:      e.CreatedAt.Unix(),
		LastActiveAt:   e.LastActiveAt.Unix(),
	}
}

// ToMap конвертирует SessionInfo в плоский map для HSET.
func (e *SessionInfo) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ip":           e.IP,
		"last_ip":      e.LastIP,
		"isp":          e.ISP,
		"country_code": e.CountryCode,
		"country_name": e.CountryName,
		"city":         e.City,
		"timezone":     e.Timezone,
		"device_type":  e.DeviceType,
		"os":           e.OS,
		"os_version":   e.OSVersion,
		"browser":      e.Browser,
		"browser_ver":  e.BrowserVersion,
		"ua":           e.UserAgentRaw,
		"created_at":   strconv.FormatInt(e.CreatedAt.Unix(), 10),
		"last_active":  strconv.FormatInt(e.LastActiveAt.Unix(), 10),
	}
}

// FromMap восстанавливает SessionInfo из результата HGETALL.
func (e *SessionInfo) FromMap(fields map[string]string) {
	e.IP = fields["ip"]
	e.LastIP = fields["last_ip"]
	e.ISP = fields["isp"]
	e.CountryCode = fields["country_code"]
	e.CountryName = fields["country_name"]
	e.City = fields["city"]
	e.Timezone = fields["timezone"]
	e.DeviceType = fields["device_type"]
	e.OS = fields["os"]
	e.OSVersion = fields["os_version"]
	e.Browser = fields["browser"]
	e.BrowserVersion = fields["browser_ver"]
	e.UserAgentRaw = fields["ua"]

	if ts, err := strconv.ParseInt(fields["created_at"], 10, 64); err == nil {
		e.CreatedAt = time.Unix(ts, 0)
	}
	if ts, err := strconv.ParseInt(fields["last_active"], 10, 64); err == nil {
		e.LastActiveAt = time.Unix(ts, 0)
	}
}
