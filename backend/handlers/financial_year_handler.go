package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/services"
)

// GetPublicFinancialYears returns all closed financial years for a unit.
// Public read — no auth. Only summary totals are exposed.
func GetPublicFinancialYears(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	svc := services.NewLedgerService()
	years, err := svc.ListYears(unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load financial years"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unitId":         unitID,
		"financialYears": years,
	})
}

// GetCurrentYearSummary returns the live summary for the current calendar year.
// Public read — same visibility as the ledger.
func GetCurrentYearSummary(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	svc := services.NewLedgerService()
	year := time.Now().UTC().Year()
	summary, err := svc.SummaryForUnitYear(unitID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unitId":  unitID,
		"year":    year,
		"summary": summary,
	})
}
