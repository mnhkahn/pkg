// Package xhex
package xhex

import (
	"bytes"
	"encoding/hex"
)

// DecodeString decodes \x like hex to bytes
func DecodeString(s string) ([]byte, error) {
	res := []byte(s)
	hhead := []byte(`\x`)
	for i := bytes.Index(res, hhead); i != -1; i = bytes.Index(res, hhead) {
		h := string(res[i+len(hhead) : i+2+len(hhead)])
		buf, err := hex.DecodeString(h)
		if err != nil {
			return nil, err
		}
		res = bytes.Replace(res, append(hhead, h...), buf, -1)
	}
	res = bytes.Replace(res, []byte(`\n`), []byte{'\n'}, -1)

	return res, nil
}
