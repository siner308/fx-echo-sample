package repository

import (
	"sync"
	"time"

	"fxserver/modules/user/entity"
)

type memoryUserRepository struct {
	users  map[int]*entity.User
	emails map[string]int
	nextID int
	mu     sync.RWMutex
}

func NewMemoryUserRepository() UserRepository {
	return &memoryUserRepository{
		users:  make(map[int]*entity.User),
		emails: make(map[string]int),
		nextID: 1,
	}
}

func (r *memoryUserRepository) Create(u *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.emails[u.Email]; exists {
		return ErrUserExists
	}

	u.ID = r.nextID
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	r.users[u.ID] = u
	r.emails[u.Email] = u.ID
	r.nextID++

	return nil
}

func (r *memoryUserRepository) GetByID(id int) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	return u, nil
}

func (r *memoryUserRepository) GetByEmail(email string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.emails[email]
	if !exists {
		return nil, ErrUserNotFound
	}

	return r.users[id], nil
}

func (r *memoryUserRepository) Update(u *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.users[u.ID]
	if !exists {
		return ErrUserNotFound
	}

	// Check if email is being changed and if new email already exists
	if u.Email != existing.Email {
		if _, emailExists := r.emails[u.Email]; emailExists {
			return ErrUserExists
		}
		delete(r.emails, existing.Email)
		r.emails[u.Email] = u.ID
	}

	u.UpdatedAt = time.Now()
	u.CreatedAt = existing.CreatedAt
	r.users[u.ID] = u

	return nil
}

func (r *memoryUserRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.users[id]
	if !exists {
		return ErrUserNotFound
	}

	delete(r.users, id)
	delete(r.emails, u.Email)

	return nil
}

func (r *memoryUserRepository) List() ([]*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*entity.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}

	return users, nil
}
