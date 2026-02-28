package logic

// Copyright 2016 Andre Burgaud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Electronic Code Book (ECB) mode.

// Implemented for legacy purpose only. ECB should be avoided
// as a mode of operation. Favor other modes available
// in the Go crypto/cipher package (i.e. CBC, GCM, CFB, OFB or CTR).

// See NIST SP 800-38A, pp 9

// The source code in this file is a modified copy of
// https://golang.org/src/crypto/cipher/cbc.go
// and released under the following
// Go Authors copyright and license:

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at https://golang.org/LICENSE

// Package ecb implements block cipher mode of encryption ECB (Electronic Code
// Book) functions. This is implemented for legacy purposes only and should not
// be used for any new encryption needs. Use CBC (Cipher Block Chaining) instead.
import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

type ecb struct {
	b         cipher.Block
	blockSize int
	tmp       []byte
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
		tmp:       make([]byte, b.BlockSize()),
	}
}

type ecbEncrypter ecb

// NewECBEncrypter returns a BlockMode which encrypts in elecronic codebook (ECB)
// mode, using the given Block (Cipher).
func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {

	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}

	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}

	for len(src) > 0 {
		x.b.Encrypt(dst[:x.blockSize], src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecrypter ecb

// NewECBDecrypter returns a BlockMode which decrypts in electronic codebook (ECB)
// mode, using the given Block.
func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}

func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	if len(src) == 0 {
		return
	}

	for len(src) > 0 {
		x.b.Decrypt(dst[:x.blockSize], src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}

}

// 项目使用特殊使用的decrypter 剥离了interface的继承使用法 为了要错误信息返回好处理
func specialDecrypter(x *ecbDecrypter, dst, src []byte) error {
	if len(src)%x.blockSize != 0 {
		return errors.New("crypto/cipher: input not full blocks")
		//panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		return errors.New("crypto/cipher: output smaller than input")
		//panic("crypto/cipher: output smaller than input")
	}
	if len(src) == 0 {
		return nil
	}

	for len(src) > 0 {
		x.b.Decrypt(dst[:x.blockSize], src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}

	return nil
}

// Copyright 2016 Andre Burgaud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package padding provides functions for padding blocks of plain text in the
// context of block cipher mode of encryption like ECB or CBC.

// Padding interface defines functions Pad and Unpad implemented for PKCS #5 and
// PKCS #7 types of padding.
type Padding interface {
	Pad(p []byte) ([]byte, error)
	Unpad(p []byte) ([]byte, error)
}

// Padder struct embeds attributes necessary for the padding calculation
// (e.g. block size). It implements the Padding interface.
type Padder struct {
	blockSize int
}

// NewPkcs5Padding returns a PKCS5 padding type structure. The blocksize
// defaults to 8 bytes (64-bit).
// See https://tools.ietf.org/html/rfc2898 PKCS #5: Password-Based Cryptography.
// Specification Version 2.0
func NewPkcs5Padding() Padding {
	return &Padder{blockSize: 8}
}

// NewPkcs7Padding returns a PKCS7 padding type structure. The blocksize is
// passed as a parameter.
// See https://tools.ietf.org/html/rfc2315 PKCS #7: Cryptographic Message
// Syntax Version 1.5.
// For example the block size for AES is 16 bytes (128 bits).
func NewPkcs7Padding(blockSize int) Padding {
	return &Padder{blockSize: blockSize}
}

// Pad returns the byte array passed as a parameter padded with bytes such that
// the new byte array will be an exact multiple of the expected block size.
// For example, if the expected block size is 8 bytes (e.g. PKCS #5) and that
// the initial byte array is:
//
//	[]byte{0x0A, 0x0B, 0x0C, 0x0D}
//
// the returned array will be:
//
//	[]byte{0x0A, 0x0B, 0x0C, 0x0D, 0x04, 0x04, 0x04, 0x04}
//
// The value of each octet of the padding is the size of the padding. If the
// array passed as a parameter is already an exact multiple of the block size,
// the original array will be padded with a full block.
func (p *Padder) Pad(buf []byte) ([]byte, error) {
	bufLen := len(buf)
	padLen := p.blockSize - (bufLen % p.blockSize)
	padText := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(buf, padText...), nil
}

// Unpad removes the padding of a given byte array, according to the same rules
// as described in the Pad function. For example if the byte array passed as a
// parameter is:
//
//	[]byte{0x0A, 0x0B, 0x0C, 0x0D, 0x04, 0x04, 0x04, 0x04}
//
// the returned array will be:
//
//	[]byte{0x0A, 0x0B, 0x0C, 0x0D}
func (p *Padder) Unpad(buf []byte) ([]byte, error) {
	bufLen := len(buf)
	if bufLen == 0 {
		return nil, errors.New("crypto/padding: invalid padding size")
	}

	pad := buf[bufLen-1]
	if pad == 0 {
		return nil, errors.New("crypto/padding: invalid last byte of padding")
	}

	padLen := int(pad)
	if padLen > bufLen || padLen > p.blockSize {
		return nil, errors.New("crypto/padding: invalid padding size")
	}

	for _, v := range buf[bufLen-padLen : bufLen-1] {
		if v != pad {
			return nil, errors.New("crypto/padding: invalid padding")
		}
	}

	return buf[:bufLen-padLen], nil
}

// aes pkcs7 加密
func AesPkcs7Encrypt(pt, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
		//panic(err.Error())
	}
	mode := NewECBEncrypter(block)
	padder := NewPkcs7Padding(mode.BlockSize())
	pt, err = padder.Pad(pt) // padd last block of plaintext if block size less than block cipher size
	if err != nil {
		return nil, err
		//panic(err.Error())
	}
	ct := make([]byte, len(pt))
	mode.CryptBlocks(ct, pt)
	return ct, nil
}

// aes pkcs7 解码
func AesPkcs7Decrypt(ct, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
		//panic(err.Error())
	}
	mode := NewECBDecrypter(block)
	pt := make([]byte, len(ct))
	//mode.CryptBlocks(pt, ct)
	err = specialDecrypter(mode.(*ecbDecrypter), pt, ct)
	padder := NewPkcs7Padding(mode.BlockSize())
	pt, err = padder.Unpad(pt) // unpad plaintext after decryption
	if err != nil {
		return nil, err
		//panic(err.Error())
	}
	return pt, nil
}
