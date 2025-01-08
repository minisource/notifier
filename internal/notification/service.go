package notification

type NotificationService struct{}

type NotificationRequest struct {
    Type    string `json:"type"`    // نوع نوتیفیکیشن (email, sms, push)
    To      string `json:"to"`      // گیرنده
    Subject string `json:"subject"` // موضوع (برای ایمیل)
    Body    string `json:"body"`    // محتوا
}

// // متد ارسال نوتیفیکیشن
// func (ns *NotificationService) SendNotification(req NotificationRequest) (string, error) {
//     switch req.Type {
//     case "email":
//         return email.SendEmail(req.To, req.Subject, req.Body)
//     case "sms":
//         return sms.SendSMS(req.To, req.Body)
//     case "push":
//         return push.SendPushNotification(req.To, req.Body)
//     default:
//         return "", errors.New("unsupported notification type")
//     }
// }