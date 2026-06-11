package logic

import "context"

type EncodingFunc func(ctx context.Context, plaintext []byte) (ciphertext []byte, err error)

type DecodingFunc func(ctx context.Context, ciphertext []byte) (plaintext []byte, err error)
