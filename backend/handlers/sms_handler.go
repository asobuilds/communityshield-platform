package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// ============================================
// SMS COMMAND STRUCTURES
// ============================================

type SMSCommand struct {
	Command   string
	Params    []string
	Phone     string
	RawInput  string
	Timestamp time.Time
}

// ============================================
// INCOMING SMS HANDLER
// ============================================

func HandleIncomingSMS(c *gin.Context) {
	// Africa's Talking / Twilio webhook format
	var input struct {
		From    string `json:"from" binding:"required"`
		Message string `json:"message" binding:"required"`
		To      string `json:"to"` // The number the SMS was sent to
		Id      string `json:"id"` // Provider's message ID
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Clean and validate input
	input.Message = strings.TrimSpace(input.Message)
	input.From = sanitizePhone(input.From)

	// Validate phone number
	if !isValidPhone(input.From) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number"})
		return
	}

	// Parse the command
	cmd := parseSMSCommand(input.Message, input.From)

	// Log incoming SMS
	log.Printf("📩 SMS Received: From=%s, Command=%s, Params=%v",
		input.From, cmd.Command, cmd.Params)

	// Process command
	response := processSMSCommand(cmd)

	// Send response back via SMS (async)
	if response != "" {
		go services.SendSMS(input.From, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "SMS processed",
		"response": response,
	})
}

// ============================================
// USSD HANDLER (Africa's Talking Format)
// ============================================

func HandleUSSD(c *gin.Context) {
	// Africa's Talking USSD format
	var input struct {
		SessionID string `json:"sessionId"`
		Phone     string `json:"phoneNumber"`
		Text      string `json:"text"`
		Network   string `json:"networkCode"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Clean input
	input.Phone = sanitizePhone(input.Phone)
	if !isValidPhone(input.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number"})
		return
	}

	// Parse USSD session
	parts := strings.Split(input.Text, "*")
	step := len(parts)

	log.Printf("📱 USSD Request: Phone=%s, Step=%d, Text=%s",
		input.Phone, step, input.Text)

	var response string

	switch step {
	case 1: // First menu
		response = handleUSSDMainMenu(input.Phone, input.SessionID)

	case 2: // Sub-menu selection
		selection := parts[0]
		response = handleUSSDSubMenu(input.Phone, selection, input.SessionID)

	default: // Data input
		response = handleUSSDData(input.Phone, input.Text, input.SessionID)
	}

	// Send response
	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

// ============================================
// USSD MENU HANDLERS
// ============================================

func handleUSSDMainMenu(phone, sessionID string) string {
	return `CON CommunityShield
1. Report Incident
2. SOS Emergency
3. Case Status
4. Help

Reply with your choice:`
}

func handleUSSDSubMenu(phone, selection, sessionID string) string {
	switch selection {
	case "1":
		return `CON Enter incident details in this format:
Title|Description|Location

Example:
Theft|Stolen phone|Lagos`
	case "2":
		// Create SOS immediately
		go createEmergencyFromUSSD(phone)
		return `END SOS alert sent! Help is on the way.`
	case "3":
		return `CON Enter your Case ID:`
	case "4":
		return `END CommunityShield Commands:
REPORT|title|desc|location
SOS|description
STATUS|case_id
HELP
Call 112 for police`
	default:
		return `END Invalid choice. Please try again.`
	}
}

func handleUSSDData(phone, text, sessionID string) string {
	// User is entering data
	if strings.HasPrefix(text, "REPORT|") || strings.Contains(text, "|") {
		return handleUSSDReport(text, phone)
	}

	// User entered a case ID
	caseID := strings.TrimSpace(text)
	if len(caseID) > 0 {
		return handleUSSDStatus(caseID, phone)
	}

	return `END Invalid input. Please try again.`
}

// ============================================
// SMS COMMAND PARSER
// ============================================

func parseSMSCommand(message, phone string) SMSCommand {
	// Split by pipe or space (support both formats)
	var parts []string
	if strings.Contains(message, "|") {
		parts = strings.Split(message, "|")
	} else {
		parts = strings.Fields(message)
	}

	cmd := SMSCommand{
		Phone:     phone,
		RawInput:  message,
		Timestamp: time.Now(),
	}

	if len(parts) > 0 {
		cmd.Command = strings.ToUpper(strings.TrimSpace(parts[0]))
	}

	if len(parts) > 1 {
		cmd.Params = make([]string, 0)
		for i := 1; i < len(parts); i++ {
			cmd.Params = append(cmd.Params, strings.TrimSpace(parts[i]))
		}
	}

	return cmd
}

// ============================================
// SMS COMMAND PROCESSOR
// ============================================

func processSMSCommand(cmd SMSCommand) string {
	switch cmd.Command {
	case "REPORT":
		return processReport(cmd)
	case "SOS":
		return processSOS(cmd)
	case "STATUS":
		return processStatus(cmd)
	case "HELP":
		return getHelpMessage()
	default:
		return getHelpMessage()
	}
}

// ============================================
// COMMAND: REPORT
// ============================================

func processReport(cmd SMSCommand) string {
	if len(cmd.Params) < 3 {
		return "Format: REPORT|title|description|location"
	}

	title := cmd.Params[0]
	description := cmd.Params[1]
	location := cmd.Params[2]

	// Validate inputs
	if len(title) < 3 {
		return "Title must be at least 3 characters"
	}
	if len(description) < 5 {
		return "Description must be at least 5 characters"
	}
	if len(location) < 2 {
		return "Please provide a valid location"
	}

	// Find or create user
	user, err := findOrCreateUser(cmd.Phone)
	if err != nil {
		log.Printf("❌ Failed to find/create user: %v", err)
		return "Service unavailable. Please try again."
	}

	// Check for duplicate reports (spam protection)
	var recentCase models.Case
	if err := config.DB.Where("reported_by = ? AND created_at > ?",
		user.ID, time.Now().Add(-5*time.Minute)).First(&recentCase).Error; err == nil {
		return "Please wait 5 minutes before reporting again."
	}

	// Create the case
	caseObj := models.Case{
		Title:       title,
		Description: description + "\n\n📱 Reported via SMS from: " + cmd.Phone,
		Location:    location,
		Status:      "pending",
		Priority:    "medium",
		IsPublic:    true,
		ReportedBy:  user.ID,
	}

	if err := config.DB.Create(&caseObj).Error; err != nil {
		log.Printf("❌ Failed to create case from SMS: %v", err)
		return "Failed to report. Please try again."
	}

	// Send confirmation
	shortID := caseObj.ID.String()[:8]
	return fmt.Sprintf("✅ Case reported! ID: %s\nCheck status with: STATUS|%s", shortID, shortID)
}

// ============================================
// COMMAND: SOS
// ============================================

func processSOS(cmd SMSCommand) string {
	description := "🚨 Emergency SOS from " + cmd.Phone
	if len(cmd.Params) > 0 && cmd.Params[0] != "" {
		description = "🚨 " + cmd.Params[0] + "\nFrom: " + cmd.Phone
	}

	// Find or create user
	user, err := findOrCreateUser(cmd.Phone)
	if err != nil {
		return "Service unavailable. Please call 112."
	}

	// Check for spam (limit 1 SOS per 5 minutes)
	var recentSOS models.SOSAlert
	if err := config.DB.Where("user_id = ? AND created_at > ?",
		user.ID, time.Now().Add(-5*time.Minute)).First(&recentSOS).Error; err == nil {
		return "SOS already sent recently. Help is on the way."
	}

	// Create SOS alert
	sos := models.SOSAlert{
		UserID:      user.ID,
		Description: description,
		Status:      "pending",
		Priority:    "high",
	}

	if err := config.DB.Create(&sos).Error; err != nil {
		log.Printf("❌ Failed to create SOS from SMS: %v", err)
		return "Failed to send SOS. Please call 112."
	}

	// Notify nearest units (async)
	go notifyUnitsAboutSOS(sos, cmd.Phone)

	return "🚨 SOS sent! Help is on the way.\nYour case ID: " + sos.ID.String()[:8]
}

// ============================================
// COMMAND: STATUS
// ============================================

func processStatus(cmd SMSCommand) string {
	if len(cmd.Params) < 1 {
		return "Format: STATUS|case_id"
	}

	caseID := cmd.Params[0]

	// Try to find by ID or short ID
	var caseObj models.Case

	// Check if it's a short ID (8 chars)
	if len(caseID) >= 8 && len(caseID) <= 12 {
		if err := config.DB.Where("id::text LIKE ?", "%"+caseID+"%").First(&caseObj).Error; err != nil {
			return "Case not found. Please check the ID."
		}
	} else {
		// Try full UUID
		id, err := uuid.Parse(caseID)
		if err != nil {
			return "Invalid case ID format. Please use the 8-character ID."
		}
		if err := config.DB.First(&caseObj, "id = ?", id).Error; err != nil {
			return "Case not found. Please check the ID."
		}
	}

	// Check if user has access
	user, _ := findOrCreateUser(cmd.Phone)
	if caseObj.ReportedBy != user.ID {
		// Check if user is admin/officer
		if user.Role != "unit_admin" && user.Role != "super_admin" {
			return "You don't have permission to view this case."
		}
	}

	return fmt.Sprintf(`📋 Case Status:
ID: %s
Title: %s
Status: %s
Priority: %s
Location: %s
Updated: %s`,
		caseObj.ID.String()[:8],
		caseObj.Title,
		caseObj.Status,
		caseObj.Priority,
		caseObj.Location,
		caseObj.UpdatedAt.Format("02 Jan 2006 15:04"))
}

// ============================================
// USSD COMMAND HANDLERS
// ============================================

func handleUSSDReport(message, phone string) string {
	parts := strings.SplitN(message, "|", 4)
	if len(parts) < 4 {
		return "END Invalid format. Use: REPORT|title|description|location"
	}

	title := parts[1]
	description := parts[2]
	location := parts[3]

	user, err := findOrCreateUser(phone)
	if err != nil {
		return "END Service unavailable."
	}

	caseObj := models.Case{
		Title:       title,
		Description: description + "\n\n📱 Reported via USSD from: " + phone,
		Location:    location,
		Status:      "pending",
		Priority:    "medium",
		IsPublic:    true,
		ReportedBy:  user.ID,
	}

	if err := config.DB.Create(&caseObj).Error; err != nil {
		return "END Failed to report."
	}

	return "END ✅ Case reported! ID: " + caseObj.ID.String()[:8]
}

func handleUSSDStatus(caseID, phone string) string {
	var caseObj models.Case
	if err := config.DB.Where("id::text LIKE ?", "%"+caseID+"%").First(&caseObj).Error; err != nil {
		return "END Case not found."
	}

	return fmt.Sprintf("END Case: %s\nStatus: %s\nPriority: %s",
		caseObj.Title, caseObj.Status, caseObj.Priority)
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// Find or create user by phone number
func findOrCreateUser(phone string) (*models.User, error) {
	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err == nil {
		return &user, nil
	}

	// Create new user
	user = models.User{
		Phone:     phone,
		FirstName: "SMS",
		LastName:  "User",
		Status:    "active",
		Role:      "citizen",
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Create emergency from USSD
func createEmergencyFromUSSD(phone string) {
	user, err := findOrCreateUser(phone)
	if err != nil {
		return
	}

	sos := models.SOSAlert{
		UserID:      user.ID,
		Description: "🚨 Emergency from USSD: " + phone,
		Status:      "pending",
		Priority:    "high",
	}
	config.DB.Create(&sos)
	go notifyUnitsAboutSOS(sos, phone)
}

// Notify units about SOS
func notifyUnitsAboutSOS(sos models.SOSAlert, phone string) {
	var units []models.SecurityUnit
	config.DB.Where("status = ?", "active").Find(&units)

	// Log notification
	log.Printf("🚨 SOS Alert: %s from %s", sos.ID.String()[:8], phone)

	for _, unit := range units {
		if unit.ContactPhone != "" {
			message := fmt.Sprintf("🚨 SOS Alert!\nID: %s\nFrom: %s\n%s\nPlease respond immediately.",
				sos.ID.String()[:8], phone, sos.Description)
			go services.SendSMS(unit.ContactPhone, message)
			log.Printf("📱 Notified unit %s at %s", unit.Name, unit.ContactPhone)
		}
	}
}

// Get help message
func getHelpMessage() string {
	return `📋 CommunityShield SMS Commands:

REPORT|title|description|location - Report a case
SOS|description - Send emergency alert
STATUS|case_id - Check case status
HELP - Show this menu

Example: REPORT|Theft|Stolen at market|Lagos`
}

// ============================================
// VALIDATION FUNCTIONS
// ============================================

func sanitizePhone(phone string) string {
	// Remove any non-digit characters
	re := regexp.MustCompile(`[^0-9]`)
	phone = re.ReplaceAllString(phone, "")

	// If starts with 0, replace with 234
	if len(phone) > 0 && phone[0] == '0' {
		phone = "234" + phone[1:]
	}

	// If doesn't start with 234, add it
	if len(phone) > 0 && len(phone) >= 10 && phone[:3] != "234" {
		phone = "234" + phone
	}

	return phone
}

func isValidPhone(phone string) bool {
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}
	re := regexp.MustCompile(`^[0-9]+$`)
	return re.MatchString(phone)
}
