package services

import (
	"log"
	"regexp"
)

// SendSMS sends a generic SMS message
func SendSMS(phone, message string) error {
	phone = formatPhoneNumber(phone)
	log.Printf("📱 SMS to %s: %s", phone, message)
	return nil
}

// SendOTPSMS sends an OTP verification code via SMS
func SendOTPSMS(phone, otpCode string) error {
	phone = formatPhoneNumber(phone)
	message := "Your CommunityShield verification code is: " + otpCode + ". It expires in 10 minutes."
	log.Printf("📱 OTP SMS to %s: %s", phone, message)
	return nil
}

// formatPhoneNumber formats phone number for Nigeria
func formatPhoneNumber(phone string) string {
	re := regexp.MustCompile(`[^0-9]`)
	phone = re.ReplaceAllString(phone, "")
	
	if len(phone) > 0 && phone[0] == '0' {
		phone = "234" + phone[1:]
	}
	
	if len(phone) > 0 && len(phone) >= 10 && phone[:3] != "234" {
		phone = "234" + phone
	}
	
	return phone
}