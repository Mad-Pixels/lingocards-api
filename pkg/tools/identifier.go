package tools

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func EncodeBaseID(args ...string) string {
	hash := md5.Sum([]byte(strings.Join(args, "::")))
	return hex.EncodeToString(hash[:])
}

func DecodeBaseID(id string) []string {
	if len(id) != 32 {
		return nil
	}
	if _, err := hex.DecodeString(id); err != nil {
		return nil
	}
	return strings.Split(id, "::")
}
