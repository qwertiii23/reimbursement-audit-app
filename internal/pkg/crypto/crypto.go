package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

// AESEncrypt AES加密
func AESEncrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM模式的加密器
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// AESDecrypt AES解密
func AESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM模式的解密器
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 检查密文长度
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// 提取nonce和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// RSAEncrypt RSA加密
func RSAEncrypt(data []byte, publicKey []byte) ([]byte, error) {
	// 解析公钥
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 类型断言
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	// 加密
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, data, nil)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

// RSADecrypt RSA解密
func RSADecrypt(data []byte, privateKey []byte) ([]byte, error) {
	// 解析私钥
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 类型断言
	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	// 解密
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaPriv, data, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

// GenerateRSAKeyPair 生成RSA密钥对
func GenerateRSAKeyPair(bits int) (privateKey, publicKey []byte, err error) {
	// 生成私钥
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	// 编码私钥
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	// 编码公钥
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return privPEM, pubPEM, nil
}

// EncryptWithPassword 使用密码加密
func EncryptWithPassword(data []byte, password string) ([]byte, error) {
	// 使用密码生成密钥
	key := sha256.Sum256([]byte(password))
	
	// 使用AES加密
	return AESEncrypt(data, key[:])
}

// DecryptWithPassword 使用密码解密
func DecryptWithPassword(data []byte, password string) ([]byte, error) {
	// 使用密码生成密钥
	key := sha256.Sum256([]byte(password))
	
	// 使用AES解密
	return AESDecrypt(data, key[:])
}

// EncryptToBase64 加密并转换为Base64
func EncryptToBase64(data []byte, key []byte) (string, error) {
	encrypted, err := AESEncrypt(data, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptFromBase64 从Base64解密
func DecryptFromBase64(data string, key []byte) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return AESDecrypt(encrypted, key)
}

// Hash 哈希函数
func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashString 哈希字符串
func HashString(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// HashWithSalt 带盐哈希
func HashWithSalt(data []byte, salt []byte) []byte {
	combined := append(data, salt...)
	hash := sha256.Sum256(combined)
	return hash[:]
}

// HashStringWithSalt 带盐哈希字符串
func HashStringWithSalt(data string, salt string) string {
	combined := data + salt
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

// GenerateRandomBytes 生成随机字节
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}