package tokens

/*
JWT 로그인
- AccessToken : 인증용
- RefreshToken : accessToken 재발급용 < Redis에 저장
*/

import (
	"jwt/redis"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	_ "github.com/twinj/uuid"
)

const ACCESS_SECRET = "gojwtauthotization"

type TokenDetails struct {
	Account   string
	Token     string
	AtExpires int64
}

// AccessToken 토큰 생성 함수
func IssuanceAccessToken(account string) (*TokenDetails, error) { // 회원가입한 account로 토큰 생성 (생성된 토큰, 에러)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	atExp := time.Now().Add(time.Minute * 20).Unix() // 20분 유효
	claims["sub"] = account
	claims["exp"] = atExp
	tk, err := token.SignedString([]byte(ACCESS_SECRET))

	if err != nil {
		return &TokenDetails{}, err
	}

	td := &TokenDetails{
		Account:   account,
		Token:     tk,
		AtExpires: atExp,
	}
	return td, nil
}

// Refresh Token 생성
func IssuanceRefreshToken(account string) (*TokenDetails, error) {
	refreshToken := jwt.New(jwt.SigningMethodHS256) // 토큰 생성
	claims := refreshToken.Claims.(jwt.MapClaims)
	atExp := time.Now().Add(time.Hour * 24 * 7).Unix()
	claims["exp"] = atExp
	claims["sub"] = account

	rtk, err := refreshToken.SignedString([]byte(ACCESS_SECRET))

	if err != nil {
		return &TokenDetails{}, err
	}

	td := &TokenDetails{
		Account:   account,
		Token:     rtk,
		AtExpires: atExp,
	}

	// redis에 저장 key = account, value = refreshToken
	err = redis.SetToken(account, rtk, td.AtExpires) // key의 time.Duration = refreshToken 수명
	if err != nil {
		panic(err)
	}

	return td, nil
}

// AccessToken, refreshToken 검증
func VerifyToken(accessToken string) bool {

	if accessToken == "" { // access-token이 비어있다면
		// response error 커스텀(?)하기
		return false
	}

	claims := jwt.MapClaims{}

	// 검증된, 만료되지 않은 토큰일 경우 tok.Valid는 true 아닐경우 false의 값을 가진다.
	token, err := jwt.ParseWithClaims(accessToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // 잘못된 토큰일 경우
			return nil, errors.New("Unexpected Siging Method.") // secret 코드 제공X
		}
		return []byte(ACCESS_SECRET), nil // secret 코드 제공O
	})

	if err != nil {
		// 에러 처리 로직 추가
		return false
	}

	return token.Valid
}

// refresh-token을 이용한 access-token 재발급
// access-token 재발급 시 refresh_token도 재발급 -> refreshToken 탈취 등 때문에
func ReissuanceAccessToken(refreshToken string) (*TokenDetails, *TokenDetails, error) {
	var accessTokenDetail *TokenDetails
	var refreshTokenDetail *TokenDetails

	if refreshToken == "" { // refresh-token이 비어있다면
		// 에러 처리
		return &TokenDetails{}, &TokenDetails{}, errors.New("Token ReIssuance Fail.")
	}

	claims := jwt.MapClaims{}

	// 검증된, 만료되지 않은 토큰일 경우 token.Valid는 true 아닐경우 false의 값을 가진다.
	token, err := jwt.ParseWithClaims(refreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // 잘못된 토큰일 경우
			return nil, errors.New("Unexpected Siging Method.") // secret 코드 제공X
		}
		return []byte(ACCESS_SECRET), nil // secret 코드 제공O
	})

	if err != nil {
		return &TokenDetails{}, &TokenDetails{}, err
	}

	// 검증된, 만료되지 않은 토큰일 경우 token.Valid는 true -> 토큰 재발급
	if token.Valid {
		account := claims["sub"].(string)
		accessTokenDetail, err = IssuanceAccessToken(account) // access_token 재발급
		if err != nil {
			return &TokenDetails{}, &TokenDetails{}, err
		}
		refreshTokenDetail, err = IssuanceRefreshToken(account) // refresh_token 재발급
		if err != nil {
			return &TokenDetails{}, &TokenDetails{}, err
		}
		err := redis.DeleteRefreshToken(refreshToken)
		if err != nil {
			return &TokenDetails{}, &TokenDetails{}, err
		}
	}

	return accessTokenDetail, refreshTokenDetail, nil
}

func TokenAuthMiddleware() fiber.Handler {
	// 회원가입과 로그인은 Token 인증이 필요없음
	return func(c *fiber.Ctx) error {
		url := c.Path()
		if FilterURL(url) {
			c.Next()
			return nil
		} else {
			// BlackList(로그아웃) 사용자 처리
			accessToken := c.Get("access-token")
			if redis.IsBlackList(accessToken) {
				return errors.New("BlackList User.")
			}
			// access-token 헤더가 비어있는 경우
			if accessToken == "" {
				return c.JSON(fiber.Map{
					"message":      "토큰 인증 실패",
					"access-token": accessToken,
				})
			}

			if VerifyToken(accessToken) { // access-token 유효
				return c.Next()
			} else { // access-token 만료
				cookieRefreshToken := string(c.Request().Header.Cookie("refresh-token")) // refresh-token 가져오기

				if VerifyToken(cookieRefreshToken) { // refresh-token 유효
					account := ExtractSubFromClaims(cookieRefreshToken)
					// redis에서 refreshToken 가져오기
					refreshToken := redis.GetRefreshToken(account)

					// access-token 재발급
					accessTokentokenDetail, refreshTokentokenDetail, err := ReissuanceAccessToken(refreshToken)

					if err != nil {
						return err
					}

					// 쿠키 설정
					c.Cookie(&fiber.Cookie{
						Name:     "refresh-token",
						Value:    refreshTokentokenDetail.Token,
						HTTPOnly: true,
						Expires:  time.Unix(refreshTokentokenDetail.AtExpires, 0),
					})

					return c.JSON(fiber.Map{
						"message":       "access-token, refreshToken 재발급",
						"access-token":  accessTokentokenDetail.Token,
						"refresh-token": refreshTokentokenDetail.Token,
					})

				} else { // access-token 만료, refreshToken 만료
					return c.JSON(fiber.Map{
						"message":       "재로그인",
						"access-token":  nil,
						"refresh-token": nil,
					})
				}
			} // end of else
		}
	}

}

// 토큰에서 claims에서 sub 추출
func ExtractSubFromClaims(refreshToken string) string {

	claims := jwt.MapClaims{}

	jwt.ParseWithClaims(refreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // 잘못된 토큰일 경우
			return nil, errors.New("Unexpected Siging Method.") // secret 코드 제공X
		}
		return []byte(ACCESS_SECRET), nil // secret 코드 제공O
	})

	return claims["sub"].(string)

}

// 토큰 검증 미들웨어 적용하지 않을 URL
func FilterURL(url string) bool {
	result := false // 변수명 넘 고민
	switch url {
	case "/api/v1/signup":
		result = true
	case "/api/v1/login":
		result = true
	case "/api/v1/logout": // logout URL을 여기다가 추가하는게 맞나
		result = true
	}
	return result
}
