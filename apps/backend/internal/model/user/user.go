package user

import "time"

type UserAccount struct {
	ID              string    `json:"id" db:"id"`
	Email           string    `json:"email" db:"email"`
	DisplayName     string    `json:"display_name" db:"display_name"`
	AvatarURL       string    `json:"avatar_url,omitempty" db:"avatar_url"`
	DefaultCurrency string    `json:"default_currency" db:"default_currency"`
	Timezone        string    `json:"timezone" db:"timezone"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
