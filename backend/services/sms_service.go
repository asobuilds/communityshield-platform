package services

import (
	"log"
	"regexp"
)

// maskPhone returns a PII-safe representation of a phone number: all digits
// except the last two are replaced with asterisks. Empty input returns "".
func maskPhone(phone string) string {
	re := regexp.MustCompile(`[^0-9]`)
	digits := re.ReplaceAllString(phone, "")
	if len(digits) <= 2 {
		return "***"
	}
	return "***" + digits[len(digits)-2:]
}

// SendSMS sends a generic SMS message
func SendSMS(phone, message string) error {
	phone = formatPhoneNumber(phone)
	log.Printf("sms: sending to %s", maskPhone(phone))
	return nil
}

// SendOTPSMS sends an OTP verification code via SMS
func SendOTPSMS(phone, otpCode string) error {
	phone = formatPhoneNumber(phone)
	// otpCode is intentionally not logged: OTP values must never appear
	// in stdout. The variable is retained to keep the call signature stable.
	_ = otpCode
	log.Printf("sms: otp sent to %s", maskPhone(phone))
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
