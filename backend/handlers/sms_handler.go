package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func InitSMS() error {
	username := os.Getenv("AT_USERNAME")
	apiKey := os.Getenv("AT_API_KEY")
	if username == "" || apiKey == "" {
		return fmt.Errorf("Africa's Talking credentials not set")
	}
	return nil
}

func SendSMS(to, message string) error {
	username := os.Getenv("AT_USERNAME")
	apiKey := os.Getenv("AT_API_KEY")
	if username == "" || apiKey == "" {
		return fmt.Errorf("AT credentials missing")
	}
	if to == "" || message == "" {
		return fmt.Errorf("recipient and message are required")
	}

	if to[:1] != "+" {
		to = "+234" + to
	}

	shortcode := os.Getenv("AT_SHORTCODE")
	if shortcode == "" {
		shortcode = "12345"
	}

	url := "https://api.africastalking.com/version1/messaging"
	payload := fmt.Sprintf("username=%s&to=%s&message=%s&from=%s",
		username, to, message, shortcode)

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return fmt.Errorf("SMS API error: %s - %s", resp.Status, string(body))
	}

	return nil
}

func SendPasswordResetSMS(to, resetLink string) error {
	message := fmt.Sprintf("CommunityShield: Reset your password using this link: %s", resetLink)
	return SendSMS(to, message)
}

func SendSOSAlertSMS(to, unitName, message string) error {
	sms := fmt.Sprintf("🚨 SOS Alert: %s\nMessage: %s", unitName, message)
	if len(sms) > 160 {
		sms = sms[:157] + "..."
	}
	return SendSMS(to, sms)
}

func SendCaseStatusSMS(to, caseTitle, status string) error {
	sms := fmt.Sprintf("Case '%s' status updated to %s. View at: http://localhost:5173/my-cases", caseTitle, status)
	if len(sms) > 160 {
		sms = sms[:157] + "..."
	}
	return SendSMS(to, sms)
}

func SendUnitRegistrationSMS(to, unitName string) error {
	sms := fmt.Sprintf("Welcome to CommunityShield! Your unit '%s' is registered. Login: http://localhost:5173/login", unitName)
	return SendSMS(to, sms)
}

// SendOTPSMS sends an OTP code via SMS
func SendOTPSMS(to, code string) error {
	message := fmt.Sprintf("CommunityShield: Your OTP code is %s. Valid for 10 minutes.", code)
	return SendSMS(to, message)
}