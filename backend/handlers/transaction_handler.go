package handlers

import (
	"net/http"
	"security-solution/models"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateTransaction(c *gin.Context) {
	var input struct {
		UnitID          string  `json:"unitId" binding:"required"`
		Amount          float64 `json:"amount" binding:"required"`
		Type            string  `json:"type" binding:"required"`
		Description     string  `json:"description" binding:"required"`
		TransactionDate string  `json:"transactionDate"`
		PaymentMethod   string  `json:"paymentMethod"`
		InitiatedBy     string  `json:"initiatedBy" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unitID, err := uuid.Parse(input.UnitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	initiatedBy, err := uuid.Parse(input.InitiatedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid initiatedBy ID"})
		return
	}

	var txnDate time.Time
	if input.TransactionDate == "" {
		txnDate = time.Now()
	} else {
		txnDate, err = time.Parse("2006-01-02", input.TransactionDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
			return
		}
	}

	txn := models.Transaction{
		UnitID:          unitID,
		Amount:          input.Amount,
		Type:            input.Type,
		Description:     input.Description,
		TransactionDate: txnDate,
		InitiatedBy:     initiatedBy,
		Status:          "pending",
		PaymentMethod:   input.PaymentMethod,
	}

	if err := DB.Create(&txn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Transaction created",
		"transaction": txn,
	})
}

func GetUnitTransactions(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedUnitID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var txns []models.Transaction
	if err := DB.Where("unit_id = ?", parsedUnitID).Order("transaction_date desc").Find(&txns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": txns})
}

func ApproveTransaction(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	var input struct {
		ApprovedBy string `json:"approvedBy" binding:"required"`
		Status     string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approvedBy, err := uuid.Parse(input.ApprovedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid approver ID"})
		return
	}

	var txn models.Transaction
	if err := DB.First(&txn, "id = ?", parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	txn.Status = input.Status
	txn.ApprovedBy = &approvedBy
	if err := DB.Save(&txn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transaction updated",
		"transaction": txn,
	})
}

func GetUnitReport(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedUnitID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	period := c.DefaultQuery("period", "monthly")
	var startDate time.Time
	now := time.Now()
	switch period {
	case "daily":
		startDate = now.AddDate(0, 0, -1)
	case "weekly":
		startDate = now.AddDate(0, 0, -7)
	default:
		startDate = now.AddDate(0, -1, 0)
	}

	var txns []models.Transaction
	if err := DB.Where("unit_id = ? AND transaction_date >= ?", parsedUnitID, startDate).
		Find(&txns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	var totalIncome, totalExpenses float64
	for _, t := range txns {
		if t.Type == "expense" {
			totalExpenses += t.Amount
		} else {
			totalIncome += t.Amount
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"period":        period,
		"startDate":     startDate,
		"totalIncome":   totalIncome,
		"totalExpenses": totalExpenses,
		"balance":       totalIncome - totalExpenses,
		"transactions":  txns,
	})
}

// NEW: GetUnitStatement for date range
func GetUnitStatement(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedUnitID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	start := c.Query("start")
	end := c.Query("end")
	if start == "" || end == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end dates are required"})
		return
	}
	startDate, err := time.Parse("2006-01-02", start)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format, use YYYY-MM-DD"})
		return
	}
	endDate, err := time.Parse("2006-01-02", end)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format, use YYYY-MM-DD"})
		return
	}
	// set end date to end of day
	endDate = endDate.Add(24*time.Hour - time.Second)

	var txns []models.Transaction
	if err := DB.Where("unit_id = ? AND transaction_date BETWEEN ? AND ?", parsedUnitID, startDate, endDate).
		Order("transaction_date asc").Find(&txns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	var totalIncome, totalExpenses float64
	for _, t := range txns {
		if t.Type == "expense" {
			totalExpenses += t.Amount
		} else {
			totalIncome += t.Amount
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"startDate":     startDate.Format("2006-01-02"),
		"endDate":       endDate.Format("2006-01-02"),
		"totalIncome":   totalIncome,
		"totalExpenses": totalExpenses,
		"balance":       totalIncome - totalExpenses,
		"transactions":  txns,
	})
}