package main

import "errors"

func SliceRemoveString(s []string, x string) ([]string, error) {
	for i, v := range s {
		if v == x {
			return append(s[0:i], s[i+1:len(s)]...), nil
		}
	}
	return s, errors.New("input slice does not contain value to remove")
}

func Uint32(x uint32) *uint32 {
	return &x
}
