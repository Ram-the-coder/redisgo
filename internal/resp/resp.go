// Redis serialization protocol (RESP)
package resp

import "fmt"

func SimpleString(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func Array(n int) string {
	return fmt.Sprintf("*%d\r\n", n)
}

func BulkString(str string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}
