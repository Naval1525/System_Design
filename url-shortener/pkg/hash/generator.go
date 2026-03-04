package base62

const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Encode(num int64) string {

	if num == 0 {
		return "0"
	}

	result := ""

	for num > 0 {

		remainder := num % 62
		result = string(charset[remainder]) + result
		num = num / 62
	}

	return result
}
