package structs

type Queue struct {
	ID int `json:"id"`
	Name string `json:"name"`
	UserID string `json:"user_id"`
	WebhookUrl string `json:"webhook_url"`
}