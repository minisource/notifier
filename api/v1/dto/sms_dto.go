package dto

type SMSRequest struct {
	To       string `json:"to"`       // گیرنده
	Body     string `json:"body"`     // محتوا
	Template string `json:"template"` // محتوا
}
