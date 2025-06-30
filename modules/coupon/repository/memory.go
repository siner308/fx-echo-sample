package repository

import (
	"fxserver/modules/coupon/entity"
	"sync"
	"time"
)

type memoryCouponRepository struct {
	coupons map[int]*entity.Coupon
	codes   map[string]int
	nextID  int
	mu      sync.RWMutex
}

func NewMemoryCouponRepository() CouponRepository {
	return &memoryCouponRepository{
		coupons: make(map[int]*entity.Coupon),
		codes:   make(map[string]int),
		nextID:  1,
	}
}

func (r *memoryCouponRepository) Create(c *entity.Coupon) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.codes[c.Code]; exists {
		return ErrCouponExists
	}

	c.ID = r.nextID
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	// Set initial status if not set
	if c.Status == "" {
		if c.IsExpired() {
			c.Status = entity.CouponStatusExpired
		} else {
			c.Status = entity.CouponStatusActive
		}
	}

	r.coupons[c.ID] = c
	r.codes[c.Code] = c.ID
	r.nextID++

	return nil
}

func (r *memoryCouponRepository) GetByID(id int) (*entity.Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	coupon, exists := r.coupons[id]
	if !exists {
		return nil, ErrCouponNotFound
	}

	return coupon, nil
}

func (r *memoryCouponRepository) GetByCode(code string) (*entity.Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.codes[code]
	if !exists {
		return nil, ErrCouponNotFound
	}

	return r.coupons[id], nil
}

func (r *memoryCouponRepository) Update(c *entity.Coupon) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.coupons[c.ID]
	if !exists {
		return ErrCouponNotFound
	}

	// Check if code is being changed and if new code already exists
	if c.Code != existing.Code {
		if _, codeExists := r.codes[c.Code]; codeExists {
			return ErrCouponExists
		}
		delete(r.codes, existing.Code)
		r.codes[c.Code] = c.ID
	}

	c.UpdatedAt = time.Now()
	c.CreatedAt = existing.CreatedAt
	r.coupons[c.ID] = c

	return nil
}

func (r *memoryCouponRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	coupon, exists := r.coupons[id]
	if !exists {
		return ErrCouponNotFound
	}

	delete(r.coupons, id)
	delete(r.codes, coupon.Code)

	return nil
}

func (r *memoryCouponRepository) List() ([]*entity.Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	coupons := make([]*entity.Coupon, 0, len(r.coupons))
	for _, coupon := range r.coupons {
		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

func (r *memoryCouponRepository) ListByStatus(status entity.CouponStatus) ([]*entity.Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var coupons []*entity.Coupon
	for _, coupon := range r.coupons {
		if coupon.Status == status {
			coupons = append(coupons, coupon)
		}
	}

	return coupons, nil
}