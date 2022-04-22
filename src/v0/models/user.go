package models

type User struct {
	ID        string `gorm:"primaryKey" json:"id"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	PFPUrl    string `json:"pfp_url" binding:"required"`
}
