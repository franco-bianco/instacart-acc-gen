package gen

import (
	"fmt"
	"time"

	"github.com/k0kubun/pp"
)

func (s *Session) GenAccount() error {
	s.Log.Info("generating account...")

	if err := s.setProxy(); err != nil {
		return fmt.Errorf("failed to set proxy: %w", err)
	}

	if err := s.createCapSolverTask(); err != nil {
		return fmt.Errorf("failed to create capSolver task: %w", err)
	}

	if err := s.getCaptchaSolution(); err != nil {
		return fmt.Errorf("failed to get captcha solution: %w", err)
	}

	if err := s.createAccount(); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	if err := s.applyValidCoupon(); err != nil {
		return fmt.Errorf("failed to apply coupon: %w", err)
	}

	account := &NewAccount{
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		Email:       s.state.Email,
		Password:    s.state.Password,
		CouponValue: s.state.CouponValue,
	}

	if err := account.AddEntry(); err != nil {
		return fmt.Errorf("failed to add account entry: %w", err)
	}

	pp.Println(account)
	s.Log.Info("successfully generated account!")

	return nil
}
