package handler

import (
	"encoding/json"
	"jwt/models"
	"jwt/redis"
	UserRepo "jwt/repository"
	"jwt/tokens"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

// 에러 처리 고민.. 에러 반환 or 에러 반환X 두루뭉실 500 Response 넘기고 에러는 로그로 남길지 고민
// 회원가입
func SignUp(c *fiber.Ctx) error {
	var user models.User

	if err := json.Unmarshal(c.Body(), &user); err != nil {
		return errors.New("SignUp Fail.")
	}
	// 아이디 중복 확인
	if _, err := UserRepo.FindByAccount(user.Account); err != nil {
		return HttpResponse(c, 409, "Duplicated Account.")
	}
	if err := user.HashPassword(user.Password); err != nil {
		return errors.New("Password Encoding Fail.")
	}

	if err := UserRepo.Save(&user); err != nil {
		return errors.New("SignUpUser Save Fail.")
	}

	return HttpResponse(c, 200, "Successfully SignUp")
}

// DTO는 OOP의 패턴입니다. Go는 객체 지향적이지 않으므로 객체를 사용하지 않기 때문에 DTO를 사용하지 않습니다.
// Go는 디자인 성보다 단순성을 포용합니다.

// 로그인 -> accesstoken, refreshToken 발급
func Login(c *fiber.Ctx) error {
	var user models.User
	var loginUser map[string]string

	if err := json.Unmarshal(c.Body(), &loginUser); err != nil {
		return errors.New("JSON Conversion Fail.")
	}

	// user 테이블에서 user 조회.
	if ValidUser, err := UserRepo.FindByAccount(loginUser["account"]); err != nil {
		return HttpResponse(c, 401, "Invalid Account")
	} else {
		user = *ValidUser
	}
	// 비밀번호 확인
	if err := user.CheckPassword(loginUser["password"]); err != nil {
		return HttpResponse(c, 401, "Invalid Password")
	}

	// Access-Token 발급
	accessTokenDetail, err := tokens.IssuanceAccessToken(loginUser["account"])
	if err != nil {
		return errors.New("AccessToken Issuance Fail.")
		//return HttpResponse(c, 500, "AccessToken Issuance Failed.")
	}

	// refreshToken 발급
	refreshTokenDetail, err := tokens.IssuanceRefreshToken(loginUser["account"])
	if err != nil {
		return errors.New("RefreshToken Issuance Fail.")
		//return HttpResponse(c, 500, "RefreshToken Issuance Failed.")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh-token",
		Value:    refreshTokenDetail.Token,
		HTTPOnly: true,
		Expires:  time.Unix(refreshTokenDetail.AtExpires, 0),
	})

	return TokenResponse(c, 200, "Successfully Login", accessTokenDetail.Token)
}

// 로그아웃
func LogOut(c *fiber.Ctx) error {
	// accesstoken 초기화 > blackList 처리
	accessToken := c.Get("access-token")
	// accessToken 남은 유효기간만큼 redis 유효기간 설정해야함... > 고민중
	// 일단은 남은 유효기간이 아닌 20분으로 설정해두고 20분 지나면 초기화되도록 함.
	// key : access-toekn, value : blackList
	err := redis.SetToken(accessToken, "blackList", time.Now().Add(time.Minute*20).Unix()) // blackList 처리

	if err != nil {
		return errors.New("AccessToken BlackList Fail.")
	}
	// refreshToken 초기화
	refreshToken := c.Request().Header.Cookie("refresh-token")
	account := tokens.ExtractSubFromClaims(string(refreshToken))

	if account == "" {
		return errors.New("Token Extract Fail.")
	}

	err = redis.DeleteRefreshToken(account)

	if err != nil {
		return errors.New("RefreshToken Delete Fail.")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh-token",
		Value:    "",
		HTTPOnly: true,
		Expires:  time.Now(),
	})
	return HttpResponse(c, 200, "Successfully Logout")
}
