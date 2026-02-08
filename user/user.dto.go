package user

import "github.com/google/uuid"

type UpdateProfileRequest struct {
	Username *string `json:"username,omitempty"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	Balance      float64   `json:"balance"`
	IsActive     bool      `json:"is_active"`
	IsRestricted bool      `json:"is_restricted"`
	CreatedAt    string    `json:"created_at"`
}

type UserFilter struct {
	Username *string `form:"username"`
	Role     *string `form:"role"`
	IsActive *bool   `form:"is_active"`
	Limit    int     `form:"limit,default=20"`
	Offset   int     `form:"offset,default=0"`
}

type SendFriendRequestRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

type FriendResponse struct {
	ID        uuid.UUID     `json:"id"`
	UserID    uuid.UUID     `json:"user_id"`
	FriendID  uuid.UUID     `json:"friend_id"`
	Status    string        `json:"status"`
	User      *UserResponse `json:"user,omitempty"`
	Friend    *UserResponse `json:"friend,omitempty"`
	CreatedAt string        `json:"created_at"`
}
