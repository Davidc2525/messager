package user

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (this *User) String() string {
	return "User(id: " + this.Id + ", Name: " + this.Name + ")"
}
