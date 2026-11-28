package entity

type ActiveUser struct {
	User          User
	ActiveAvatars int
	LastAvatar    Avatar
}
