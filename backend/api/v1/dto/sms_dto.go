package dto

type SMSRequest struct {
	To          string `json:"to"`          // گیرنده
	PhoneNumber string `json:"phoneNumber"` // شماره تلفن
	Body        string `json:"body"`        // محتوا
	Template    string `json:"template"`
}
