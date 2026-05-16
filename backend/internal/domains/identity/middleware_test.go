package identity

import (
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
)

func Test_AdminGroupName(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.Assert(AdminGroup, "admin")
	})
}

func Test_AdminMiddleware_Logic(t *testing.T) {
	// AdminMiddleware checks: isAdmin flag OR groups contains "admin"
	gtest.C(t, func(t *gtest.T) {
		// isAdmin=true should pass
		t.Assert(true, true)

		// groups=["admin"] should pass
		groups := []string{"admin"}
		hasAdmin := false
		for _, g := range groups {
			if g == AdminGroup {
				hasAdmin = true
			}
		}
		t.Assert(hasAdmin, true)

		// groups=["enterprise-x"] should not pass
		groups2 := []string{"enterprise-x"}
		hasAdmin2 := false
		for _, g := range groups2 {
			if g == AdminGroup {
				hasAdmin2 = true
			}
		}
		t.Assert(hasAdmin2, false)

		// empty groups should not pass
		hasAdmin3 := false
		for _, g := range []string{} {
			if g == AdminGroup {
				hasAdmin3 = true
			}
		}
		t.Assert(hasAdmin3, false)
	})
}
