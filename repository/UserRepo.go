package repository

import (
	"jwt/db"
	"jwt/models"
)

// 회원 추가
func Save(user *models.User) error {
	if err := db.PostgreDB.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// 회원 조회
func FindByAccount(account string) (*models.User, error) {
	var user models.User

	if err := db.PostgreDB.Find(&user, "account = ?", account).Error; err != nil {
		return &models.User{}, err
	}
	// for _, u := range UserRepo {
	// 	if u.Account == account {
	// 		user = u
	// 	}
	// }
	return &user, nil
}
