package common

import "fmt"

func ParseUintParam(p string) uint {
	var id uint
	fmt.Sscanf(p, "%d", &id)
	return id
}
