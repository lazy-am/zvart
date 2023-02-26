package cipher

import (
	"fmt"
	"testing"

	"github.com/lazy-am/zvart/pkg/random"
)

func TestAES(t *testing.T) {
	for i := 1; i < 30; i++ {
		testname := fmt.Sprintf("Test aes, string len %d", i)
		t.Run(testname, func(t *testing.T) {
			v := random.RandStringBytes(i)
			key := GetSHA256([]byte(random.RandStringBytes(10)))
			buf, err := AESEncript(key, []byte(v))
			if err != nil {
				t.Fatal("encript error")
			}
			res, err := AESDecript(key, buf)
			if err != nil {
				t.Fatal("decript error")
			}
			if string(res) != v {
				t.Fatalf("strings are not equal, result is %s", string(res))
			}
		})
	}

}
