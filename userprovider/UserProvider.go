package userprovider

import (
	"fmt"
	"github.com/Davidc2525/messager/user"
	//"github.com/Davidc2525/test_signleton/userprovider"
)

var instance *UserProvider

type UserProvider struct {
	users map[string]*user.User
}

func (this *UserProvider) GetUserById(id string) (*user.User, error) {
	if user, ok := this.users[id]; !ok {
		return nil, fmt.Errorf("No existe usuario %s", id)
	} else {
		return user, nil
	}
}

//Crear nuevo usuario
func (this *UserProvider) AddUser(u *user.User) (user *user.User) {
	this.users[u.Id] = u
	user = u
	return
}

func (this *UserProvider) GetUsers() map[string]*user.User {
	return this.users
}

func GetInstance() *UserProvider {
	if instance == nil {
		instance = &UserProvider{make(map[string]*user.User)}
	}
	return instance
}
