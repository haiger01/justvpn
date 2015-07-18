/*
* @Author: BlahGeek
* @Date:   2015-06-28
* @Last Modified by:   BlahGeek
* @Last Modified time: 2015-07-18
 */

package obfs

import "fmt"

type XorObfusecator struct {
	key     []byte
	max_len int
}

func (xor *XorObfusecator) Open(options map[string]interface{},
	max_obfsed_len int) error {
	if field := options["key"]; field == nil {
		return fmt.Errorf("`key` not found in options")
	} else {
		key_str := field.(string)
		xor.key = make([]byte, len(key_str))
		copy(xor.key, key_str)
	}
	xor.max_len = max_obfsed_len
	return nil
}

func (xor *XorObfusecator) Close() error { return nil }

func (xor *XorObfusecator) GetMaxPlainLength() int { return xor.max_len }

func (xor *XorObfusecator) Encode(src, dst []byte) int {
	for i := 0; i < len(src); i += 1 {
		c := xor.key[i%len(xor.key)]
		dst[i] = c ^ src[i]
	}
	return len(src)
}

// For xor, decoding is same as encoding, error is always nil
func (xor *XorObfusecator) Decode(src, dst []byte) (int, error) {
	return xor.Encode(src, dst), nil
}
