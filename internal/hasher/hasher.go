package hasher

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct{}

/*
*

	Hash(password string) ([]byte, error)
	Compare(hash []byte, password string) (bool, error)
*/
func (b *BcryptHasher) Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (b *BcryptHasher) Compare(hash []byte, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))

	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
