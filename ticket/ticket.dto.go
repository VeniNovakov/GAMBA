package ticket

import "github.com/google/uuid"

type CreateRequest struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Priority    string `json:"priority,omitempty"` // defaults to "medium"
}

type UpdateRequest struct {
	Status     *string    `json:"status,omitempty"`
	Priority   *string    `json:"priority,omitempty"`
	AssignedTo *uuid.UUID `json:"assigned_to,omitempty"`
}

type AddMessageRequest struct {
	Content string `json:"content"`
}
