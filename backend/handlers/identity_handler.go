package handlers

import (
    "crypto/sha256"
    "encoding/hex"
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "security-solution/config"
    "security-solution/models"
)

type identityVerificationRequest struct {
    DocumentType   string `json:"documentType" binding:"required"`
    DocumentNumber string `json:"documentNumber" binding:"required"`
    DocumentURL    string `json:"documentUrl"`
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
    value, exists := c.Get("user_id")
    if !exists {
        return uuid.Nil, false
    }

    switch v := value.(type) {
    case uuid.UUID:
        return v, true
    case string:
        id, err := uuid.Parse(v)
        if err != nil {
            return uuid.Nil, false
        }
        return id, true
    default:
        return uuid.Nil, false
    }
}

// SubmitIdentityVerification creates or replaces a pending identity
// verification request for the authenticated user.
func SubmitIdentityVerification(c *gin.Context) {
    userID, ok := currentUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    var req identityVerificationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    req.DocumentType = strings.TrimSpace(req.DocumentType)
    req.DocumentNumber = strings.TrimSpace(req.DocumentNumber)
    req.DocumentURL = strings.TrimSpace(req.DocumentURL)

    if req.DocumentType == "" || req.DocumentNumber == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "document type and document number are required",
        })
        return
    }

    hash := sha256.Sum256([]byte(req.DocumentNumber))
    documentHash := hex.EncodeToString(hash[:])

    var existing models.IdentityVerification
    err := config.DB.Where("user_id = ?", userID).First(&existing).Error

    if err == nil {
        if existing.Status == "verified" {
            c.JSON(http.StatusConflict, gin.H{
                "error": "identity is already verified",
            })
            return
        }

        existing.DocumentType = req.DocumentType
        existing.DocumentNumberHash = documentHash
        existing.DocumentURL = req.DocumentURL
        existing.Status = "pending"
        existing.RejectionReason = ""
        existing.VerifiedBy = nil
        existing.VerifiedAt = nil
        existing.SubmittedAt = time.Now()

        if err := config.DB.Save(&existing).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "failed to submit identity verification",
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "identity verification resubmitted",
            "verification": gin.H{
                "id":     existing.ID,
                "status": existing.Status,
            },
        })
        return
    }

    verification := models.IdentityVerification{
        ID:                 uuid.New(),
        UserID:             userID,
        DocumentType:       req.DocumentType,
        DocumentNumberHash: documentHash,
        DocumentURL:        req.DocumentURL,
        Status:             "pending",
        SubmittedAt:        time.Now(),
    }

    if err := config.DB.Create(&verification).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "failed to submit identity verification",
        })
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "identity verification submitted",
        "verification": gin.H{
            "id":     verification.ID,
            "status": verification.Status,
        },
    })
}

// GetMyIdentityVerification returns only the authenticated user's
// non-sensitive verification state.
func GetMyIdentityVerification(c *gin.Context) {
    userID, ok := currentUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    var verification models.IdentityVerification

    if err := config.DB.
        Where("user_id = ?", userID).
        First(&verification).Error; err != nil {

        c.JSON(http.StatusOK, gin.H{
            "status": "not_submitted",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":             verification.ID,
        "status":         verification.Status,
        "documentType":   verification.DocumentType,
        "submittedAt":    verification.SubmittedAt,
        "verifiedAt":     verification.VerifiedAt,
        "rejectionReason": verification.RejectionReason,
    })
}
