package media

func Transform(message []byte) []int16 {
	var result []int16

	length := len(message)
	step := 2
	shiftBits := (step - 1) * 8
	for index := 0; index < length - step + 1; index += step {
		result = append(result, int16(message[index + 1]) << shiftBits + int16(message[index]))
	}
	return result
}
