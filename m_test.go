package main

import (
	"fmt"
	"monitor/utils"
	"strings"
	"testing"
)

func TestMoveSpace(t *testing.T) {
	str := utils.MoveMoreSpace("tcp        0      0 127.0.0.1:43879         0.0.0.0:*               LISTEN      1227/node")
	// fmt.Println(str)
	splits := strings.Split(str, " ")

	lastIndex := strings.LastIndex(splits[3], ":")
	if lastIndex != -1 {
		fmt.Println("lastIndex: ", splits[3][lastIndex+1:])
	}
}
