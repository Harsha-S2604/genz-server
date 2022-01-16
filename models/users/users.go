package users

// use promoted fields if possible
type User struct {
	UserId 			string `json: "userId"`
	Name			string `json: "name"`
	Email 			string `json: "email"`
	Password 		string `json: "password"`
	Profile			string `json: "profile"`
}

func New(userId string, name string, email string, password string, profile string) User {
	return User{userId, name, email, password, profile}
}

func (u *User) GetUserId() string {
	return u.UserId
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetPassword() string {
	return u.Password
}

func (u *User) SetUserId(userId string) {
	u.UserId = userId
}

func (u *User) SetName(name string) {
	u.Name = name
}

func (u *User) SetEmail(email string) {
	u.Email = email
}

func (u *User) SetPassword(password string) {
	u.Password = password
}