package format

import "time"

// TimePtr форматирует указатель на time.Time в строку RFC3339 (UTC).
// Возвращает пустую строку, если указатель nil.
func TimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// UnixTimestamp форматирует Unix-метку (секунды) в строку RFC3339 (UTC).
// Возвращает пустую строку, если ts == 0.
func UnixTimestamp(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}
