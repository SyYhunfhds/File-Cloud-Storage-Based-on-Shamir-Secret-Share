package logic

import (
	"context"

	"github.com/gogf/gf/v2/encoding/gbase64"
)

// 后处理系列函数

// PEFBase64Encode 后处理函数: 对数据进行base64编码
var PEFBase64Encode EncodingFunc = func(ctx context.Context, plaintext []byte) (ciphertext []byte, err error) {
	return gbase64.Encode(plaintext), nil
}

var PDFBase64Decode DecodingFunc = func(ctx context.Context, ciphertext []byte) (plaintext []byte, err error) {
	return gbase64.Decode(ciphertext)
}

func ChainEncode(ctx context.Context, plaintext []byte, processes ...EncodingFunc) (ciphertext []byte, err error) {
	for _, fn := range processes {
		plaintext, err = fn(ctx, plaintext)
		if err != nil {
			return nil, err
		}
	}
	return plaintext, nil
}
func ChainDecode(ctx context.Context, ciphertext []byte, processes ...DecodingFunc) (plaintext []byte, err error) {
	for _, fn := range processes {
		ciphertext, err = fn(ctx, ciphertext)
		if err != nil {
			return nil, err
		}
	}
	return ciphertext, nil
}
