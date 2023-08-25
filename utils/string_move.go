package utils

import "bytes"

// 移除字符串中多余的空格
func MoveMoreSpace(origin string) string {
	buf := []byte(origin)

	writer := bytes.NewBuffer([]byte{})
	space_index := -1
	get_letter := false

	for i := 0; i < len(buf); i++ {
		if buf[i] == ' ' {
			if space_index == -1 && get_letter {
				writer.WriteByte(buf[i])
				space_index = i
			}
		} else {
			get_letter = true
			if space_index != -1 {
				space_index = -1
			}
			writer.WriteByte(buf[i])
		}
	}
	return writer.String()
}
