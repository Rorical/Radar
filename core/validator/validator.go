package validator

import (
	"fmt"
)

type Validator struct {
}

func (v Validator) Validate(key string, value []byte) error {
	fmt.Println(key)
	return nil
}

func (v Validator) Select(key string, values [][]byte) (int, error) {
	fmt.Println(key)
	return 0, nil
}
