package handlers

import (
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/services"
)

// GetHealth returns the current health status of the server and its
// subsystems. It is PUBLIC (no auth) so external monitors can poll it.
//
// HTTP status: 200 when all checks pass or only the scheduler is stale
// (status="degraded"); 503 when the database is unreachable
// (status="down").
func GetHealth(c *gin.Context) {
	now := time.Now()
	uptime := now.Sub(services.ProcessStart()).Seconds()

	version := os.Getenv("BUILD_VERSION")
	if version == "" {
		version = "dev"
	}

	checks := map[string]map[string]interface{}{}

	// --- database check: ping with 2s timeout, fail-closed ---
	dbOK, dbDetail := checkDatabase()
	checks["database"] = map[string]interface{}{
		"ok":    dbOK,
		"detail": dbDetail,
	}

	// --- scheduler check: boot grace + staleness ---
	schedOK, schedDetail, schedStale := checkScheduler(now, uptime)
	checks["scheduler"] = map[string]interface{}{
		"ok":    schedOK,
		"detail": schedDetail,
	}

	// --- overall status ---
	overall := "ok"
	if !dbOK {
		overall = "down"
	} else if schedStale {
		overall = "degraded"
	}

	statusCode := 200
	if overall == "down" {
		statusCode = 503
	}

	c.JSON(statusCode, gin.H{
		"status":        overall,
		"uptimeSeconds": uptime,
		"version":       version,
		"timestamp":     now.UTC().Format(time.RFC3339Nano),
		"checks":        checks,
	})
}

// checkDatabase pings the underlying sql.DB with a 2s timeout.
func checkDatabase() (bool, string) {
	sqlDB, err := config.DB.DB()
	if err != nil {
		return false, "cannot get sql.DB: " + err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return false, "ping failed: " + err.Error()
	}
	return true, "ok"
}

// checkScheduler evaluates the scheduler health. Boot grace: during the
// first 900s the scheduler may not have run yet, so we report ok with a
// "boot grace" detail. After that, if the last run is more than 2h stale
// we report a degraded status (not down — the server can still serve
// traffic). Returns (ok, detail, isStale).
func checkScheduler(now time.Time, uptime float64) (bool, string, bool) {
	const bootGrace = 900.0   // 15 minutes
	const staleThreshold = 2 * time.Hour

	if uptime < bootGrace {
		return true, "boot grace", false
	}

	lastRun := services.SchedulerLastRunAt()
	if lastRun.IsZero() {
		return false, "scheduler has never run", true
	}
	if now.Sub(lastRun) > staleThreshold {
		return false, "scheduler stale", true
	}
	return true, "ok", false
}

// GetMetrics returns in-process counters. Super-admin only: non-super
// authed users get 403, unauthenticated get 401.
func GetMetrics(c *gin.Context) {
	snap := services.SnapshotMetrics()
	c.JSON(200, gin.H{
		"requestsTotal":    snap.RequestsTotal,
		"requestsByStatus": gin.H{
			"2xx": snap.Requests2xx,
			"3xx": snap.Requests3xx,
			"4xx": snap.Requests4xx,
			"5xx": snap.Requests5xx,
		},
		"errorsTotal":      snap.ErrorsTotal,
		"averageLatencyMs": snap.AverageLatencyMs,
		"uptimeSeconds":    snap.UptimeSeconds,
	})
}