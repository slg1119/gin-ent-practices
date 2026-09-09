package user

import "context"

type UserService struct {
	userRepository UserRepository
}

func NewUserService(userRepository UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) Create(ctx context.Context, name, email string) (*User, error) {
	u, err := NewUser(name, email)
	if err != nil {
		return nil, err
	}
	// 중복 여부는 DB의 unique 제약으로 보장한다. 사전 조회만으로는 동시 가입을 막을 수 없다.
	return s.userRepository.Create(ctx, u)
}

func (s *UserService) Get(ctx context.Context, id int) (*User, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	return s.userRepository.FindByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]*User, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, ErrInvalidPagination
	}
	return s.userRepository.List(ctx, limit, offset)
}

func (s *UserService) Deactivate(ctx context.Context, id int) (*User, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	return s.userRepository.Deactivate(ctx, id)
}
