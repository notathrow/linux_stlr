package browsers

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"errors"
	"fmt"
	"hash"
)

func PBKDF2Key(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	u := make([]byte, hashLen)
	for block := 1; block <= numBlocks; block++ {
		// N.B.: || means concatenation, ^ means XOR
		// for each block T_i = U_1 ^ U_2 ^ ... ^ U_iter
		// U_1 = PRF(password, salt || uint(i))
		prf.Reset()
		prf.Write(salt)
		buf[0] = byte(block >> 24)
		buf[1] = byte(block >> 16)
		buf[2] = byte(block >> 8)
		buf[3] = byte(block)
		prf.Write(buf[:4])
		dk = prf.Sum(dk)
		t := dk[len(dk)-hashLen:]
		copy(u, t)

		for n := 2; n <= iter; n++ {
			prf.Reset()
			prf.Write(u)
			u = u[:0]
			u = prf.Sum(u)
			for x := range u {
				t[x] ^= u[x]
			}
		}
	}
	return dk[:keyLen]
}

var ErrCiphertextLengthIsInvalid = errors.New("ciphertext length is invalid")

func AES128CBCDecrypt(key, iv, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// Check ciphertext length
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("AES128CBCDecrypt: ciphertext too short")
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("AES128CBCDecrypt: ciphertext is not a multiple of the block size")
	}

	decryptedData := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decryptedData, ciphertext)
	// unpad the decrypted data and handle potential padding errors
	decryptedData, err = pkcs5UnPadding(decryptedData)
	if err != nil {
		return nil, fmt.Errorf("AES128CBCDecrypt: %w", err)
	}

	return decryptedData, nil
}

func AES128CBCEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != aes.BlockSize {
		return nil, errors.New("AES128CBCEncrypt: iv length is invalid, must equal block size")
	}

	plaintext = pkcs5Padding(plaintext, block.BlockSize())
	encryptedData := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptedData, plaintext)

	return encryptedData, nil
}

func DES3Decrypt(key, iv, ciphertext []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < des.BlockSize {
		return nil, errors.New("DES3Decrypt: ciphertext too short")
	}
	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("DES3Decrypt: ciphertext is not a multiple of the block size")
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	sq := make([]byte, len(ciphertext))
	blockMode.CryptBlocks(sq, ciphertext)

	return pkcs5UnPadding(sq)
}

func DES3Encrypt(key, iv, plaintext []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext = pkcs5Padding(plaintext, block.BlockSize())
	dst := make([]byte, len(plaintext))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(dst, plaintext)

	return dst, nil
}

// AESGCMDecrypt chromium > 80 https://source.chromium.org/chromium/chromium/src/+/master:components/os_crypt/os_crypt_win.cc
func AESGCMDecrypt(key, nounce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	origData, err := blockMode.Open(nil, nounce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return origData, nil
}

// AESGCMEncrypt encrypts plaintext using AES encryption in GCM mode.
func AESGCMEncrypt(key, nonce, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// The first parameter is the prefix for the output, we can leave it nil.
	// The Seal method encrypts and authenticates the data, appending the result to the dst.
	encryptedData := blockMode.Seal(nil, nonce, plaintext, nil)
	return encryptedData, nil
}

func paddingZero(src []byte, length int) []byte {
	padding := length - len(src)
	if padding <= 0 {
		return src
	}
	return append(src, make([]byte, padding)...)
}

func pkcs5UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return nil, errors.New("pkcs5UnPadding: src should not be empty")
	}
	padding := int(src[length-1])
	if padding < 1 || padding > aes.BlockSize {
		return nil, errors.New("pkcs5UnPadding: invalid padding size")
	}
	return src[:length-padding], nil
}

func pkcs5Padding(src []byte, blocksize int) []byte {
	padding := blocksize - (len(src) % blocksize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

type ASN1PBE interface {
	Decrypt(globalSalt []byte) ([]byte, error)

	Encrypt(globalSalt, plaintext []byte) ([]byte, error)
}

func NewASN1PBE(b []byte) (pbe ASN1PBE, err error) {
	var (
		nss   nssPBE
		meta  metaPBE
		login loginPBE
	)
	if _, err := asn1.Unmarshal(b, &nss); err == nil {
		return nss, nil
	}
	if _, err := asn1.Unmarshal(b, &meta); err == nil {
		return meta, nil
	}
	if _, err := asn1.Unmarshal(b, &login); err == nil {
		return login, nil
	}
	return nil, ErrDecodeASN1Failed
}

var ErrDecodeASN1Failed = errors.New("decode ASN1 data failed")

// nssPBE Struct
//
//	SEQUENCE (2 elem)
//		OBJECT IDENTIFIER
//		SEQUENCE (2 elem)
//			OCTET STRING (20 byte)
//			INTEGER 1
//	OCTET STRING (16 byte)
type nssPBE struct {
	AlgoAttr struct {
		asn1.ObjectIdentifier
		SaltAttr struct {
			EntrySalt []byte
			Len       int
		}
	}
	Encrypted []byte
}

// Decrypt decrypts the encrypted password with the global salt.
func (n nssPBE) Decrypt(globalSalt []byte) ([]byte, error) {
	key, iv := n.deriveKeyAndIV(globalSalt)

	return DES3Decrypt(key, iv, n.Encrypted)
}

func (n nssPBE) Encrypt(globalSalt, plaintext []byte) ([]byte, error) {
	key, iv := n.deriveKeyAndIV(globalSalt)

	return DES3Encrypt(key, iv, plaintext)
}

// deriveKeyAndIV derives the key and initialization vector (IV)
// from the global salt and entry salt.
func (n nssPBE) deriveKeyAndIV(globalSalt []byte) ([]byte, []byte) {
	salt := n.AlgoAttr.SaltAttr.EntrySalt
	hashPrefix := sha1.Sum(globalSalt)
	compositeHash := sha1.Sum(append(hashPrefix[:], salt...))
	paddedEntrySalt := paddingZero(salt, 20)

	hmacProcessor := hmac.New(sha1.New, compositeHash[:])
	hmacProcessor.Write(paddedEntrySalt)

	paddedEntrySalt = append(paddedEntrySalt, salt...)
	keyComponent1 := hmac.New(sha1.New, compositeHash[:])
	keyComponent1.Write(paddedEntrySalt)

	hmacWithSalt := append(hmacProcessor.Sum(nil), salt...)
	keyComponent2 := hmac.New(sha1.New, compositeHash[:])
	keyComponent2.Write(hmacWithSalt)

	key := append(keyComponent1.Sum(nil), keyComponent2.Sum(nil)...)
	iv := key[len(key)-8:]
	return key[:24], iv
}

// MetaPBE Struct
//
//	SEQUENCE (2 elem)
//		OBJECT IDENTIFIER
//	    SEQUENCE (2 elem)
//	    SEQUENCE (2 elem)
//	      	OBJECT IDENTIFIER
//	       	SEQUENCE (4 elem)
//	       	OCTET STRING (32 byte)
//	      		INTEGER 1
//	       		INTEGER 32
//	       		SEQUENCE (1 elem)
//	          	OBJECT IDENTIFIER
//	    SEQUENCE (2 elem)
//	      	OBJECT IDENTIFIER
//	      	OCTET STRING (14 byte)
//	OCTET STRING (16 byte)
type metaPBE struct {
	AlgoAttr  algoAttr
	Encrypted []byte
}

type algoAttr struct {
	asn1.ObjectIdentifier
	Data struct {
		Data struct {
			asn1.ObjectIdentifier
			SlatAttr slatAttr
		}
		IVData ivAttr
	}
}

type ivAttr struct {
	asn1.ObjectIdentifier
	IV []byte
}

type slatAttr struct {
	EntrySalt      []byte
	IterationCount int
	KeySize        int
	Algorithm      struct {
		asn1.ObjectIdentifier
	}
}

func (m metaPBE) Decrypt(globalSalt []byte) ([]byte, error) {
	key, iv := m.deriveKeyAndIV(globalSalt)

	return AES128CBCDecrypt(key, iv, m.Encrypted)
}

func (m metaPBE) Encrypt(globalSalt, plaintext []byte) ([]byte, error) {
	key, iv := m.deriveKeyAndIV(globalSalt)

	return AES128CBCEncrypt(key, iv, plaintext)
}

func (m metaPBE) deriveKeyAndIV(globalSalt []byte) ([]byte, []byte) {
	password := sha1.Sum(globalSalt)

	salt := m.AlgoAttr.Data.Data.SlatAttr.EntrySalt
	iter := m.AlgoAttr.Data.Data.SlatAttr.IterationCount
	keyLen := m.AlgoAttr.Data.Data.SlatAttr.KeySize

	key := PBKDF2Key(password[:], salt, iter, keyLen, sha256.New)
	iv := append([]byte{4, 14}, m.AlgoAttr.Data.IVData.IV...)
	return key, iv
}

// loginPBE Struct
//
//	OCTET STRING (16 byte)
//	SEQUENCE (2 elem)
//			OBJECT IDENTIFIER
//			OCTET STRING (8 byte)
//	OCTET STRING (16 byte)
type loginPBE struct {
	CipherText []byte
	Data       struct {
		asn1.ObjectIdentifier
		IV []byte
	}
	Encrypted []byte
}

func (l loginPBE) Decrypt(globalSalt []byte) ([]byte, error) {
	key, iv := l.deriveKeyAndIV(globalSalt)
	return DES3Decrypt(key, iv, l.Encrypted)
}

func (l loginPBE) Encrypt(globalSalt, plaintext []byte) ([]byte, error) {
	key, iv := l.deriveKeyAndIV(globalSalt)
	return DES3Encrypt(key, iv, plaintext)
}

func (l loginPBE) deriveKeyAndIV(globalSalt []byte) ([]byte, []byte) {
	return globalSalt, l.Data.IV
}

func DecryptWithChromium(key, encryptPass []byte) ([]byte, error) {
	if len(encryptPass) < 3 {
		return nil, ErrCiphertextLengthIsInvalid
	}
	iv := []byte{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32}
	return AES128CBCDecrypt(key, iv, encryptPass[3:])
}

func DecryptWithDPAPI(_ []byte) ([]byte, error) {
	return nil, nil
}
