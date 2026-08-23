package interlock

import (
	"fmt"

	"github.com/lacsar712/reefctl/internal/model"
)

func CheckDefrostPending() error {
	return fmt.Errorf("defrost guard: %w", model.ErrDefrostPending)
}
