package logic

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"xfx/gm_server/dto"
)

type mainServerTimeEnvelope struct {
	ErrCode int            `json:"errcode"`
	ErrMsg  string         `json:"errmsg"`
	Data    map[string]any `json:"data"`
}

func normalizeMainServerTimeResponse(serverID int64, body []byte) (int, string, *dto.GMServerTimePayload, error) {
	var envelope mainServerTimeEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return ERR_SERVER_INTERNAL, "invalid main_server response", nil, err
	}
	if envelope.ErrMsg == "" {
		envelope.ErrMsg = "success"
	}

	raw := make(map[string]any)
	for key, value := range envelope.Data {
		raw[key] = value
	}

	var topLevel map[string]any
	if err := json.Unmarshal(body, &topLevel); err == nil {
		for _, key := range []string{
			"time",
			"serverTime",
			"server_time",
			"gameTime",
			"game_time",
			"gameIso",
			"game_iso",
			"realTime",
			"real_time",
			"realIso",
			"real_iso",
			"offsetDays",
			"offset_days",
			"offsetEnabled",
			"offset_enabled",
		} {
			if _, exists := raw[key]; !exists {
				if value, ok := topLevel[key]; ok {
					raw[key] = value
				}
			}
		}
	}

	payload := buildServerTimePayload(serverID, raw)
	return envelope.ErrCode, envelope.ErrMsg, payload, nil
}

func buildServerTimePayload(serverID int64, raw map[string]any) *dto.GMServerTimePayload {
	gameTime := firstInt64(raw, "gameTime", "game_time")
	realTime := firstInt64(raw, "realTime", "real_time")
	gameIso := firstString(raw, "gameIso", "gameISO", "game_iso")
	realIso := firstString(raw, "realIso", "realISO", "real_iso")
	offsetDays := firstInt64(raw, "offsetDays", "offset_days")
	offsetEnabled := firstBool(raw, "offsetEnabled", "offset_enabled")

	timeText := firstString(raw, "time", "serverTime", "server_time")
	if timeText == "" {
		timeText = formatDisplayTime(gameIso)
	}
	if timeText == "" {
		timeText = formatUnixTime(gameTime)
	}
	if timeText == "" {
		timeText = formatDisplayTime(realIso)
	}
	if timeText == "" {
		timeText = formatUnixTime(realTime)
	}

	return &dto.GMServerTimePayload{
		ServerID:      serverID,
		Time:          timeText,
		ServerTime:    timeText,
		GameTime:      gameTime,
		GameIso:       gameIso,
		RealTime:      realTime,
		RealIso:       realIso,
		OffsetDays:    offsetDays,
		OffsetEnabled: offsetEnabled,
	}
}

func buildServerTimeLegacy(payload *dto.GMServerTimePayload) map[string]any {
	return map[string]any{
		"time":           payload.Time,
		"serverTime":     payload.ServerTime,
		"server_id":      payload.ServerID,
		"game_time":      payload.GameTime,
		"game_iso":       payload.GameIso,
		"real_time":      payload.RealTime,
		"real_iso":       payload.RealIso,
		"offset_days":    payload.OffsetDays,
		"offset_enabled": payload.OffsetEnabled,
	}
}

func firstString(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok || value == nil {
			continue
		}
		text := strings.TrimSpace(toString(value))
		if text != "" {
			return text
		}
	}
	return ""
}

func firstInt64(raw map[string]any, keys ...string) int64 {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok || value == nil {
			continue
		}
		if parsed, ok := toInt64(value); ok {
			return parsed
		}
	}
	return 0
}

func firstBool(raw map[string]any, keys ...string) bool {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok || value == nil {
			continue
		}
		if parsed, ok := toBool(value); ok {
			return parsed
		}
	}
	return false
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(value)
	}
}

func toInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case float32:
		return int64(typed), true
	case float64:
		return int64(typed), true
	case json.Number:
		v, err := typed.Int64()
		return v, err == nil
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return 0, false
		}
		v, err := strconv.ParseInt(text, 10, 64)
		return v, err == nil
	default:
		return 0, false
	}
}

func toBool(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		text := strings.TrimSpace(strings.ToLower(typed))
		switch text {
		case "1", "true", "yes", "on":
			return true, true
		case "0", "false", "no", "off":
			return false, true
		default:
			return false, false
		}
	default:
		if v, ok := toInt64(value); ok {
			return v != 0, true
		}
		return false, false
	}
}

func formatDisplayTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Local().Format("2006-01-02 15:04:05")
		}
	}
	return value
}

func formatUnixTime(value int64) string {
	if value <= 0 {
		return ""
	}
	return time.Unix(value, 0).Local().Format("2006-01-02 15:04:05")
}
