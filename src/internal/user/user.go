// Package user는 HTTP와 ORM에 의존하지 않는 회원 기능의 핵심이다.
package user

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

// User는 Ent 객체나 HTTP 응답과 별개인 도메인 모델이다.
type User struct {
	ID        int
	Name      string
	Email     string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser는 HTTP 외의 호출 경로에서도 동일한 가입 규칙을 적용한다.
func NewUser(name, email string) (*User, error) {
	name = strings.TrimSpace(name)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 100 {
		return nil, ErrInvalidName
	}

	// 이 예제에서는 이메일 전체를 대소문자 구분 없이 취급한다.
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 {
		return nil, ErrInvalidEmail
	}
	return &User{Name: name, Email: email, IsActive: true}, nil
}
