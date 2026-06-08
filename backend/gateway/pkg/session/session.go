package session

import (
	"net"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mssola/useragent"
	"github.com/oschwald/geoip2-golang"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
)

// Provider извлекает данные сессии из HTTP-запроса.
// Геолокация доступна только если при инициализации переданы пути к базам MaxMind GeoLite2.
// Если базы не найдены — гео-поля остаются пустыми, ошибки не возникает.
type Provider struct {
	cityDB *geoip2.Reader
	asnDB  *geoip2.Reader
}

// New создаёт Provider.
// cityDBPath — путь к GeoLite2-City.mmdb (пустая строка → геолокация отключена).
// asnDBPath  — путь к GeoLite2-ASN.mmdb  (пустая строка → ISP отключён).
func New(cityDBPath, asnDBPath string) *Provider {
	p := &Provider{}

	if cityDBPath != "" {
		db, err := geoip2.Open(cityDBPath)
		if err == nil {
			p.cityDB = db
		}
	}
	if asnDBPath != "" {
		db, err := geoip2.Open(asnDBPath)
		if err == nil {
			p.asnDB = db
		}
	}

	return p
}

// Close освобождает ресурсы баз данных MaxMind.
func (p *Provider) Close() {
	if p.cityDB != nil {
		_ = p.cityDB.Close()
	}
	if p.asnDB != nil {
		_ = p.asnDB.Close()
	}
}

// Extract собирает SessionInfo из текущего HTTP-запроса.
// Заполняет IP, UserAgent (устройство/ОС/браузер) и геолокацию (если база доступна).
// Поля LastIP и LastActiveAt при создании совпадают с IP и CreatedAt —
// они будут обновлены при следующем RefreshToken.
func (p *Provider) Extract(c *fiber.Ctx) *auth_proto.SessionInfo {
	now := time.Now().Unix()
	ip := ClientIP(c)
	rawUA := c.Get("User-Agent")

	deviceType, os, osVersion, browser, browserVersion := parseUA(rawUA)

	s := &auth_proto.SessionInfo{
		Ip:             ip,
		LastIp:         ip,
		DeviceType:     deviceType,
		Os:             os,
		OsVersion:      osVersion,
		Browser:        browser,
		BrowserVersion: browserVersion,
		UserAgentRaw:   rawUA,
		CreatedAt:      now,
		LastActiveAt:   now,
	}

	p.fillGeo(s, ip)
	return s
}

// ClientIP возвращает реальный IP клиента с учётом proxy-заголовков.
func ClientIP(c *fiber.Ctx) string {
	// Cloudflare
	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	// Стандартный reverse-proxy заголовок
	if fwd := c.Get("X-Forwarded-For"); fwd != "" {
		// Может содержать цепочку IP через запятую — берём первый (клиентский)
		if idx := strings.IndexByte(fwd, ','); idx != -1 {
			return strings.TrimSpace(fwd[:idx])
		}
		return strings.TrimSpace(fwd)
	}
	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return c.IP()
}

// ── Приватные хелперы ─────────────────────────────────────────────────────────

// fillGeo обогащает SessionInfo данными геолокации и ISP.
// Ничего не делает, если базы данных не инициализированы.
func (p *Provider) fillGeo(s *auth_proto.SessionInfo, rawIP string) {
	ip := net.ParseIP(rawIP)
	if ip == nil {
		return
	}

	if p.cityDB != nil {
		if record, err := p.cityDB.City(ip); err == nil {
			s.CountryCode = record.Country.IsoCode
			if name, ok := record.Country.Names["en"]; ok {
				s.CountryName = name
			}
			if name, ok := record.City.Names["en"]; ok {
				s.City = name
			}
			s.Timezone = record.Location.TimeZone
		}
	}

	if p.asnDB != nil {
		if record, err := p.asnDB.ASN(ip); err == nil {
			s.Isp = record.AutonomousSystemOrganization
		}
	}
}

// parseUA парсит строку User-Agent и возвращает (deviceType, os, osVersion, browser, browserVersion).
func parseUA(rawUA string) (deviceType, os, osVersion, browser, browserVersion string) {
	if rawUA == "" {
		return "desktop", "", "", "", ""
	}

	ua := useragent.New(rawUA)

	// Тип устройства
	lower := strings.ToLower(rawUA)
	switch {
	case strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet"):
		deviceType = "tablet"
	case ua.Mobile() || strings.Contains(lower, "mobile"):
		deviceType = "mobile"
	default:
		deviceType = "desktop"
	}

	// ОС
	os, osVersion = parseOSVersion(ua.OS())

	// Браузер
	browserName, browserVer := ua.Browser()
	browser = browserName
	browserVersion = truncateVersion(browserVer)

	return
}

// parseOSVersion разбивает строку вида "Windows 10", "iOS 17.4", "Android 11"
// на имя ОС и версию.
func parseOSVersion(raw string) (name, version string) {
	switch {
	case strings.HasPrefix(raw, "Mac OS X"):
		name = "macOS"
		parts := strings.Fields(raw)
		if len(parts) >= 4 {
			version = strings.ReplaceAll(parts[3], "_", ".")
		}
	case strings.HasPrefix(raw, "Windows"):
		name = "Windows"
		parts := strings.Fields(raw)
		if len(parts) >= 2 {
			version = parts[1]
		}
	case strings.HasPrefix(raw, "Android"):
		name = "Android"
		parts := strings.Fields(raw)
		if len(parts) >= 2 {
			version = parts[1]
		}
	case strings.HasPrefix(raw, "iOS"):
		name = "iOS"
		parts := strings.Fields(raw)
		if len(parts) >= 2 {
			version = parts[1]
		}
	default:
		parts := strings.Fields(raw)
		name = raw
		if len(parts) >= 2 {
			name = parts[0]
			version = parts[len(parts)-1]
		}
	}
	return
}

// truncateVersion обрезает версию вида "125.0.6422.112" до "125.0".
func truncateVersion(v string) string {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}
