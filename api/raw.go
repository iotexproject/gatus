package api

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/TwiN/gatus/v5/storage/store"
	"github.com/TwiN/gatus/v5/storage/store/common"
	"github.com/gofiber/fiber/v2"
)

func UptimeRaw(c *fiber.Ctx) error {
	duration := c.Params("duration")
	var from time.Time
	switch duration {
	case "30d":
		from = time.Now().Add(-30 * 24 * time.Hour)
	case "7d":
		from = time.Now().Add(-7 * 24 * time.Hour)
	case "24h":
		from = time.Now().Add(-24 * time.Hour)
	case "1h":
		from = time.Now().Add(-2 * time.Hour) // Because uptime metrics are stored by hour, we have to cheat a little
	default:
		return c.Status(400).SendString("Durations supported: 30d, 7d, 24h, 1h")
	}
	key, err := url.QueryUnescape(c.Params("key"))
	if err != nil {
		return c.Status(400).SendString("invalid key encoding")
	}
	uptime, err := store.Get().GetUptimeByKey(key, from, time.Now())
	if err != nil {
		if errors.Is(err, common.ErrEndpointNotFound) {
			return c.Status(404).SendString(err.Error())
		} else if errors.Is(err, common.ErrInvalidTimeRange) {
			return c.Status(400).SendString(err.Error())
		}
		return c.Status(500).SendString(err.Error())
	}

	c.Set("Content-Type", "text/plain")
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Expires", "0")
	return c.Status(200).Send([]byte(fmt.Sprintf("%f", uptime)))
}

// UptimeHistory returns the hourly uptime statistics of an endpoint, oldest first, for status bars spanning days
func UptimeHistory(c *fiber.Ctx) error {
	var from time.Time
	switch c.Params("duration") {
	case "30d":
		from = time.Now().Add(-30 * 24 * time.Hour)
	case "7d":
		from = time.Now().Add(-7 * 24 * time.Hour)
	default:
		return c.Status(400).SendString("Durations supported: 30d, 7d")
	}
	key, err := url.QueryUnescape(c.Params("key"))
	if err != nil {
		return c.Status(400).SendString("invalid key encoding")
	}
	hourlyStatistics, err := store.Get().GetHourlyUptimeStatisticsByKey(key, from, time.Now())
	if err != nil {
		if errors.Is(err, common.ErrEndpointNotFound) {
			return c.Status(404).SendString(err.Error())
		} else if errors.Is(err, common.ErrInvalidTimeRange) {
			return c.Status(400).SendString(err.Error())
		}
		return c.Status(500).SendString(err.Error())
	}
	type hour struct {
		Timestamp           int64  `json:"timestamp"` // Start of the hour, in milliseconds
		Total               uint64 `json:"total"`
		Successful          uint64 `json:"successful"`
		AverageResponseTime uint64 `json:"averageResponseTime"` // In milliseconds
	}
	hours := make([]hour, 0, len(hourlyStatistics))
	for timestamp, stats := range hourlyStatistics {
		hours = append(hours, hour{timestamp * 1000, stats.TotalExecutions, stats.SuccessfulExecutions, stats.TotalExecutionsResponseTime / stats.TotalExecutions})
	}
	sort.Slice(hours, func(i, j int) bool { return hours[i].Timestamp < hours[j].Timestamp })
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return c.Status(200).JSON(hours)
}

func ResponseTimeRaw(c *fiber.Ctx) error {
	duration := c.Params("duration")
	var from time.Time
	switch duration {
	case "30d":
		from = time.Now().Add(-30 * 24 * time.Hour)
	case "7d":
		from = time.Now().Add(-7 * 24 * time.Hour)
	case "24h":
		from = time.Now().Add(-24 * time.Hour)
	case "1h":
		from = time.Now().Add(-2 * time.Hour) // Because uptime metrics are stored by hour, we have to cheat a little
	default:
		return c.Status(400).SendString("Durations supported: 30d, 7d, 24h, 1h")
	}
	key, err := url.QueryUnescape(c.Params("key"))
	if err != nil {
		return c.Status(400).SendString("invalid key encoding")
	}
	responseTime, err := store.Get().GetAverageResponseTimeByKey(key, from, time.Now())
	if err != nil {
		if errors.Is(err, common.ErrEndpointNotFound) {
			return c.Status(404).SendString(err.Error())
		} else if errors.Is(err, common.ErrInvalidTimeRange) {
			return c.Status(400).SendString(err.Error())
		}
		return c.Status(500).SendString(err.Error())
	}

	c.Set("Content-Type", "text/plain")
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Expires", "0")
	return c.Status(200).Send([]byte(fmt.Sprintf("%d", responseTime)))
}
