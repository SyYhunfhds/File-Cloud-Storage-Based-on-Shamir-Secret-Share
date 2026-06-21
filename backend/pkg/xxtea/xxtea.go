package xxtea

import (
	"encoding/binary"
	"math/bits"
)

const (
	// Delta 黄金分割比例常数 (0x9E3779B9)
	Delta = 0x9E3779B9
)

// Encrypt 对 []uint32 数据进行 XXTEA 加密（原地修改）。
// n 必须 >= 2。
func Encrypt(v []uint32, key [4]uint32) []uint32 {
	n := len(v)
	if n < 2 {
		return v
	}

	rounds := 6 + 52/n
	var sum uint32
	z := v[n-1]
	for q := 0; q < rounds; q++ {
		sum += Delta
		e := (sum >> 2) & 3
		for p := 0; p < n-1; p++ {
			y := v[p+1]
			v[p] += mx(y, z, sum, key, p, e)
			z = v[p]
		}
		{
			p := n - 1
			y := v[0]
			v[p] += mx(y, z, sum, key, p, e)
			z = v[p]
		}
	}
	return v
}

// Decrypt 对 []uint32 数据进行 XXTEA 解密（原地修改）。
func Decrypt(v []uint32, key [4]uint32) []uint32 {
	n := len(v)
	if n < 2 {
		return v
	}

	rounds := 6 + 52/n
	var sum = uint32(rounds) * Delta
	y := v[0]
	for q := 0; q < rounds; q++ {
		e := (sum >> 2) & 3
		for p := n - 1; p > 0; p-- {
			z := v[p-1]
			v[p] -= mx(y, z, sum, key, p, e)
			y = v[p]
		}
		{
			z := v[n-1]
			v[0] -= mx(y, z, sum, key, 0, e)
			y = v[0]
		}
		sum -= Delta
	}
	return v
}

// EncryptUint32 对单个 uint32 进行 XXTEA 加密。
// 内部包装为 2 元素 slice 后加密。
func EncryptUint32(v uint32, key [4]uint32) uint32 {
	buf := []uint32{v, 0}
	Encrypt(buf, key)
	return buf[0]
}

// DecryptUint32 对单个 uint32 进行 XXTEA 解密。
func DecryptUint32(v uint32, key [4]uint32) uint32 {
	buf := []uint32{v, 0}
	Decrypt(buf, key)
	return buf[0]
}

// mx 混合函数 (标准 XXTEA 定义)
func mx(y, z, sum uint32, key [4]uint32, p int, e uint32) uint32 {
	return bits.RotateLeft32((z>>5 ^ y<<2)+(y>>3 ^ z<<4), int((y^sum^key[p&3^int(e)]))) ^ ((z>>5 ^ y<<2) + (y>>3 ^ z<<4))
}

// KeyFromBytes 从 16 字节的 []byte 创建 XXTEA 密钥
func KeyFromBytes(b []byte) [4]uint32 {
	if len(b) < 16 {
		bb := make([]byte, 16)
		copy(bb, b)
		b = bb
	}
	return [4]uint32{
		binary.LittleEndian.Uint32(b[0:4]),
		binary.LittleEndian.Uint32(b[4:8]),
		binary.LittleEndian.Uint32(b[8:12]),
		binary.LittleEndian.Uint32(b[12:16]),
	}
}
