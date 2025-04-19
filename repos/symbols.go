package repos

import (
	"database/sql"
	"fmt"

	"github.com/streamdp/ccd/caches"
	"github.com/streamdp/ccd/domain"
)

type SymbolsStore interface {
	AddSymbol(s string, u string) (result sql.Result, err error)
	UpdateSymbol(s string, u string) (result sql.Result, err error)
	RemoveSymbol(s string) (result sql.Result, err error)
	Symbols() (symbols []*domain.Symbol, err error)
}

type symbolRepo struct {
	s SymbolsStore
	c *caches.SymbolCache
}

func NewSymbolRepository(s SymbolsStore) *symbolRepo {
	return &symbolRepo{
		c: caches.NewSymbolCache(),
		s: s,
	}
}

func (r *symbolRepo) Update(s, u string) error {
	if _, err := r.s.UpdateSymbol(s, u); err != nil {
		return fmt.Errorf("failed to update symbol: %w", err)
	}
	r.c.Add(s)

	return nil
}

func (r *symbolRepo) Load() error {
	s, err := r.s.Symbols()
	if err != nil {
		return fmt.Errorf("failed to load symbols: %w", err)
	}
	for i := range s {
		r.c.Add(s[i].Symbol)
	}

	return nil
}

func (r *symbolRepo) GetAll() []string {
	return r.c.GetAll()
}

func (r *symbolRepo) Add(s, u string) error {
	if r.IsPresent(s) {
		return nil
	}
	if _, err := r.s.AddSymbol(s, u); err != nil {
		return fmt.Errorf("failed to add symbol: %w", err)
	}
	r.c.Add(s)

	return nil
}

func (r *symbolRepo) Remove(s string) error {
	if !r.IsPresent(s) {
		return nil
	}
	if _, err := r.s.RemoveSymbol(s); err != nil {
		return fmt.Errorf("failed to remove symbol: %w", err)
	}
	r.c.Remove(s)

	return nil
}

func (r *symbolRepo) IsPresent(s string) bool {
	return r.c.IsPresent(s)
}
