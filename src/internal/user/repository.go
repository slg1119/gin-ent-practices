package user

import "context"

// UserRepository는 서비스가 필요로 하는 저장 동작만 정의한다.
// 구현체는 Ent 타입과 DB 오류를 이 패키지의 모델과 오류로 변환한다.
type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	FindByID(ctx context.Context, id int) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	// Deactivate는 이미 비활성화된 사용자에도 성공하며, 없는 ID에는 ErrNotFound를 반환한다.
	Deactivate(ctx context.Context, id int) (*User, error)
}
