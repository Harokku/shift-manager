package auth

import "testing"

func TestComparePassword(t *testing.T) {
	var pwd = "plinioilbasso"

	hashedPwd, err := HashAndSalt(pwd)
	if err != nil {
		t.Errorf("Password hashing failed")
	}

	result := ComparePassword(hashedPwd, pwd)
	if result != true {
		t.Errorf("Hash and check failed, algorithm password mismatch")
	}
}
