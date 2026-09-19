package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetBlueprint — PUBLIC. Returns a structured description of the platform:
// its purpose, architecture, roles, capabilities, and how to use it.
// Designed to be rendered as the public /about page.
func GetBlueprint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":    "WardGuard",
		"tagline": "Community safety and governance infrastructure for Nigerian neighborhoods.",
		"purpose": "WardGuard connects citizens, security units, officers, and administrators through one auditable workflow: report → assign → dispatch → investigate → review → close. It is designed to promote accountability, prevent vigilantism, and keep sensitive case intelligence restricted to authorized personnel.",

		"principles": []gin.H{
			{"title": "Unit access is NOT case access", "body": "A security unit does not automatically see every case in its jurisdiction — only cases assigned to specific officers, or submitted to specific administrators."},
			{"title": "Every hierarchy is auditable", "body": "Head admins, admins, and officers are all removable by defined thresholds and quorum. No role is above recall."},
			{"title": "Presumed innocence", "body": "A citizen named as a suspect sees only that a case references them, its status, and public-safe progress. Never investigative detail, officer identity, or evidence internals."},
			{"title": "No vigilantism", "body": "The platform coordinates lawful response. It does not arm, direct, or encourage any citizen to confront another."},
			{"title": "Explicit consent for sensitive data", "body": "Medical information and government identity documents are encrypted at rest. Nothing auto-deletes without the user's consent, except security tombstones documented separately."},
		},

		"roles": []gin.H{
			{"name": "Citizen", "capabilities": []string{"Report incidents", "Track your case", "Rate the officer and unit after closure", "See community alerts and announcements", "Donate directly to a unit's bank account", "See if you are named in a case (self-view only)"}},
			{"name": "Officer", "capabilities": []string{"Work cases assigned to you (primary or paired)", "Support access on other cases you assist", "File weekly narratives", "Submit cases for review", "See your unit's officer leaderboard", "View your unit's financial ledger"}},
			{"name": "Unit Admin", "capabilities": []string{"Triage cases in your unit", "Assign cases to officers", "Review closures", "Manage elections and revocations", "Manage bank accounts and visibility", "Approve disbursements"}},
			{"name": "Head Admin", "capabilities": []string{"Everything a unit admin can do", "See all cases in the unit", "Override case archive", "Manage the unit's UnitAuth policy", "Publish unit-scoped news", "Delegate the treasurer role"}},
			{"name": "Super Admin", "capabilities": []string{"Platform-wide access", "Manage all units and users", "Confirm platform donations", "Audit any financial or case ledger"}},
		},

		"architecture": gin.H{
			"backend": "Go 1.24 · Gin · GORM · PostgreSQL · JWT",
			"frontend": "React 19 · TypeScript · Vite · Tailwind · React Query · Leaflet",
			"infra": "Docker Compose · Render · Supabase",
		},

		"case_lifecycle": []string{
			"pending", "assigned", "dispatched", "on_scene",
			"investigating", "pending_admin_review",
			"admin_changes_requested", "closed",
		},

		"security": []string{
			"JWT with jti-based revocation",
			"Refresh tokens with rotation",
			"Rate limiting on auth, OTP, votes, and invites",
			"Password changes revoke all sessions immediately",
			"Tiered case access (primary / paired / support)",
			"Public-safe DTOs on all citizen-facing endpoints",
			"Encrypted MedicalInfo at rest (AES-256-GCM)",
			"HMAC-signed SMS webhooks",
		},

		"finance": gin.H{
			"unit_model": "Units publish their own bank account details. Donors transfer directly — WardGuard never touches the money.",
			"ledger":     "Every donation and expense is stamped with a permanent, unique reference (WG-YYYY-NNNNNNNN). Entries are append-only.",
			"yearly_cycle": "Each December, the year closes with a frozen snapshot: total in, total out, net, top categories.",
			"platform_donation": "WardGuard itself accepts donations for maintenance via a separate public endpoint.",
		},

		"governance": gin.H{
			"elections": gin.H{
				"term_months":           12,
				"staggered_rotation":    true,
				"stagger_months":        6,
				"term_limit":            2,
				"cooling_off_months":    6,
				"seat_bands":            gin.H{"12-20": 5, "21-40": 7, "41-70": 9, "71-100": 10},
				"quorum_percent":        50,
			},
			"revocation": gin.H{
				"regular_admin": "10+ verified member votes + head admin approval",
				"head_admin":    "majority of verified unit members",
				"one_vote_per":  "per member per cycle",
			},
		},

		"ratings": gin.H{
			"method": "Bayesian smoothing",
			"tiers":  []string{"unranked", "bronze", "silver", "gold", "platinum", "diamond"},
			"rules": []string{
				"Only the reporter may rate",
				"Only after the case closes via the platform",
				"30-day window after closure",
				"One rating per case per ratee",
				"Officer leaderboards are within-unit only",
				"Officer of the Week is announced via platform news",
			},
		},

		"how_to_use": []gin.H{
			{"step": "1. Register", "detail": "Create a citizen account. No unit required."},
			{"step": "2. Report", "detail": "File an incident with location and optional evidence links."},
			{"step": "3. Track", "detail": "Watch your case move through the lifecycle, see weekly updates shared by the assigned officer."},
			{"step": "4. Rate", "detail": "After closure, rate the officer and the unit."},
			{"step": "5. Join or host a unit", "detail": "Accept a unit invite, or host your own and grow it to 12+ members to activate governance."},
		},

		"contact": gin.H{
			"support_email": "support@wardguard.app",
			"platform_url":  "https://wardguard.app",
		},
	})
}
