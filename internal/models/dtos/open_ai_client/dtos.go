package open_ai_client

import (
	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"github.com/jamesphm04/splose-clone-be/internal/types"
)

type Attachment struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Message struct {
	Role        string       `json:"role"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments"`
}

type Note struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Patient struct {
	Email       string          `json:"email"`
	FirstName   string          `json:"firstName"`
	LastName    string          `json:"lastName"`
	PhoneNumber string          `json:"phoneNumber"`
	DateOfBirth *types.Date     `json:"dateOfBirth"`
	Gender      entities.Gender `json:"gender"`
	FullAddress string          `json:"fullAddress"`
}

type ConversationContext struct {
	Patient      Patient   `json:"patient"`
	Note         Note      `json:"note"`
	Conversation []Message `json:"conversation"`
}

type SendMessageRequest struct {
	ConversationContext ConversationContext `json:"conversationContext"`
	Message             string              `json:"message"`
}

type SendMessageResponse struct {
	Message string `json:"message"`
}
