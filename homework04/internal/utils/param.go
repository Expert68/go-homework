package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func QueryInt(c *gin.Context, name string, def int) int {
	v, err := strconv.Atoi(c.Query(name))
	if err != nil {
		return def
	}
	return int(v)
}
