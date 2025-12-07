package models

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique;not null" json:"username"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

// UserEntity describes the behavior of a user model that can be used across layers
// without tying code to the concrete struct. The current concrete type User
// implements this interface via the getters and setters below.
type UserEntity interface {
	GetID() uint
	GetUsername() string
	GetRole() string
	// PasswordHash returns the stored password hash (never plain text).
	PasswordHash() string
	// SetPasswordHash sets the stored password hash.
	SetPasswordHash(hash string)
}

func (u *User) GetID() uint { return u.ID }

func (u *User) GetUsername() string { return u.Username }

func (u *User) GetRole() string { return u.Role }

func (u *User) PasswordHash() string { return u.Password }

func (u *User) SetPasswordHash(hash string) { u.Password = hash }
