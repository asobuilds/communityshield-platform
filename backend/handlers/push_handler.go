package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"security-solution/models"
	"strings"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var vapidPublicKey, vapidPrivateKey, vapidSubject string

func initVAPID() {
	vapidPublicKey = os.Getenv("VAPID_PUBLIC_KEY")
	vapidPrivateKey = os.Getenv("VAPID_PRIVATE_KEY")
	vapidSubject = os.Getenv("VAPID_SUBJECT")

	if vapidPublicKey == "" || vapidPrivateKey == "" {
		privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
		if err != nil {
			fmt.Println("❌ Failed to generate VAPID keys:", err)
			return
		}
		vapidPublicKey = publicKey
		vapidPrivateKey = privateKey
		fmt.Println("🔑 VAPID Public Key:", publicKey)
		fmt.Println("🔑 VAPID Private Key:", privateKey)
	}
	if vapidSubject == "" {
		vapidSubject = "mailto:communityshield@example.com"
	}
}

// SubscribePush saves a push subscription
func SubscribePush(c *gin.Context) {
	var input struct {
		UserID   string `json:"userId" binding:"required"`
		Endpoint string `json:"endpoint" binding:"required"`
		Keys     struct {
			P256dh string `json:"p256dh" binding:"required"`
			Auth   string `json:"auth" binding:"required"`
		} `json:"keys" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}

	keysJSON, err := json.Marshal(input.Keys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal keys"})
		return
	}

	var sub models.PushSubscription
	if err := DB.Where("user_id = ? AND endpoint = ?", userID, input.Endpoint).First(&sub).Error; err == nil {
		sub.Keys = string(keysJSON)
		if err := DB.Save(&sub).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Subscription updated"})
		return
	}

	sub = models.PushSubscription{
		UserID:   userID,
		Endpoint: input.Endpoint,
		Keys:     string(keysJSON),
	}
	if err := DB.Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save subscription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Subscribed successfully"})
}

// UnsubscribePush removes a push subscription
func UnsubscribePush(c *gin.Context) {
	userIDStr := c.Query("userId")
	endpoint := c.Query("endpoint")
	if userIDStr == "" || endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId and endpoint required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}
	if err := DB.Where("user_id = ? AND endpoint = ?", userID, endpoint).Delete(&models.PushSubscription{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}

// SendPushNotification sends a push notification to a user
func SendPushNotification(userID uuid.UUID, title, body, icon, url string) error {
	if vapidPublicKey == "" || vapidPrivateKey == "" {
		initVAPID()
	}

	var subs []models.PushSubscription
	if err := DB.Where("user_id = ?", userID).Find(&subs).Error; err != nil {
		return err
	}
	if len(subs) == 0 {
		return nil // no subscriptions
	}

	payload := map[string]string{
		"title": title,
		"body":  body,
		"icon":  icon,
		"url":   url,
	}
	payloadBytes, _ := json.Marshal(payload)

	for _, sub := range subs {
		var keys struct {
			P256dh string `json:"p256dh"`
			Auth   string `json:"auth"`
		}
		if err := json.Unmarshal([]byte(sub.Keys), &keys); err != nil {
			continue
		}

		resp, err := webpush.SendNotification(
			payloadBytes,
			&webpush.Subscription{
				Endpoint: sub.Endpoint,
				Keys: webpush.Keys{
					P256dh: keys.P256dh,
					Auth:   keys.Auth,
				},
			},
			&webpush.Options{
				Subscriber:      vapidSubject,
				VAPIDPublicKey:  vapidPublicKey,
				VAPIDPrivateKey: vapidPrivateKey,
				TTL:             60 * 60 * 24 * 7, // 7 days
			},
		)
		if err != nil {
			if strings.Contains(err.Error(), "410") || strings.Contains(err.Error(), "expired") {
				DB.Delete(&sub)
			}
			continue
		}
		defer resp.Body.Close()
	}
	return nil
}

// GetVAPIDPublicKey returns the public key for the frontend
func GetVAPIDPublicKey(c *gin.Context) {
	if vapidPublicKey == "" {
		initVAPID()
	}
	c.JSON(http.StatusOK, gin.H{"publicKey": vapidPublicKey})
}