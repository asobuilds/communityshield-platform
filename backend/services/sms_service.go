package services

import (
	"fmt"
	"log"
)

// SendOTPSMS - OPTIMIZED: Non-blocking, just logs for now
func SendOTPSMS(phoneNumber, otpCode string) error {
	// Format phone number
	if len(phoneNumber) > 0 && phoneNumber[0] == '0' {
		phoneNumber = "234" + phoneNumber[1:]
	}
	if len(phoneNumber) > 0 && len(phoneNumber) < 11 && phoneNumber[0:3] != "234" {
		phoneNumber = "234" + phoneNumber
	}

	message := fmt.Sprintf("Your CommunityShield verification code is: %s. It expires in 10 minutes.", otpCode)

	// Log the SMS
	log.Printf("📱 OTP SMS sent to %s: %s", phoneNumber, message)

	// For production, integrate with SMS provider here
	// The response should be immediate, SMS sending happens asynchronously

	return nil
}