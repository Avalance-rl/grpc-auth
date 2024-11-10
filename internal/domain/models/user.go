package models

import "time"

type User struct {
	ID               uint64    `db:"id"`
	Email            string    `db:"email"`
	PassHash         []byte    `db:"password"`
	RegistrationTime time.Time `db:"registration_time"`
}

func (u User) ReceiveNotification(sprintf string) {
	panic("implement me")
}
