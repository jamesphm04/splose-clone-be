package dtos

import (
	"time"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
)

type MessageDTO struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func ToDTO(message *entities.Message) *MessageDTO {
	return &MessageDTO{
		ID:        message.ID,
		Role:      string(message.Role),
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}
}
