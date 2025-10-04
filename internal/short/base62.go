package short

import "errors"

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func EncodeBase62(number int64) string {
	if number == 0 {
		return "0"
	}
	var encoded []byte
	n := number
	for n > 0 {
		remainder := n % 62
		encoded = append([]byte{base62Alphabet[remainder]}, encoded...)
		n = n / 62
	}
	return string(encoded)
}

func DecodeBase62(code string) (int64, error) {
	var result int64
	for i := 0; i < len(code); i++ {
		ch := code[i]
		var value int64 = -1

		switch {
		case ch >= '0' && ch <= '9':
			value = int64(ch - '0') // 0..9 -> 0..9
		case ch >= 'A' && ch <= 'Z':
			value = int64(ch-'A') + 10 // A..Z -> 10..35
		case ch >= 'a' && ch <= 'z':
			value = int64(ch-'a') + 36 // a..z -> 36..61
		}

		if value == -1 {
			return 0, errors.New("invalid base62 character")
		}
		result = result*62 + value
	}
	return result, nil
}
