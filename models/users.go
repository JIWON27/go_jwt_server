package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

/*
TODO
- error 커스텀
*/
type User struct {
	gorm.Model        // ID, CratedAt, UpdatedAt, detetedAt
	Account    string `json:"account"`
	Password   string `json:"password"`
}

// 비밀번호 암호화
func (user *User) HashPassword(password string) error {
	hashpwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil
	}
	user.Password = string(hashpwd)
	return nil
}

// 비밀번호 확인
func (user *User) CheckPassword(loginPwd string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginPwd)) // 비밀번호 일치하면 nil 반환
	if err != nil {
		return err
	}
	return nil
}
