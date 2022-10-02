package utils

import (
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-base32"
)

func DecodeB32Cid(b32str string) (cid.Cid, error) {
	byts, err := base32.RawStdEncoding.DecodeString(b32str)
	if err != nil {
		return cid.Undef, err
	}
	_, id, err := cid.CidFromBytes(byts)
	return id, err
}
