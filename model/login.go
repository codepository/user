package model

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/jinzhu/gorm"
)

var logintable = "fznews_login"

// FznewsLogin 登陆表
type FznewsLogin struct {
	Model
	UserID   int    `json:"user_id"`
	Password string `json:"password"`
}

// SaveIfNotExist 不存在就保存，存在就啥也不做
func (l *FznewsLogin) SaveIfNotExist() error {
	if l.UserID == 0 {
		return errors.New("user_id不能为空")
	}
	l.Password = encodePass(l.UserID, l.Password)
	return db.Table(logintable).Where(FznewsLogin{UserID: l.UserID}).Assign(l).FirstOrCreate(l).Error
}

// Update 只会更新这些更改的和非空白字段
func (l *FznewsLogin) Update() error {
	if l.UserID == 0 {
		return errors.New("user_id不能为空")
	}
	l.Password = encodePass(l.UserID, l.Password)
	return db.Table(logintable).Model(&FznewsLogin{}).Where("user_id=?", l.UserID).Updates(l).Error
}

// Find Find
func (l *FznewsLogin) Find() error {
	if l.UserID == 0 {
		return errors.New("user_id不能为空")
	}
	return db.Table(logintable).Where(l).Find(&l).Error
}

// Check 验证密码并生成token
func (l *FznewsLogin) Check() (bool, string, error) {
	l.Password = encodePass(l.UserID, l.Password)
	var count int
	err := db.Table(logintable).Where(l).Count(&count).Error
	if err != nil {
		return false, "", err
	}
	token := l.Password
	return count > 0, token, nil
}

// GetToken 获取token
func (l *FznewsLogin) GetToken() (string, error) {
	if l.UserID == 0 || len(l.Password) == 0 {
		return "", errors.New("user_id和password不能为空")
	}
	return encodePass(l.UserID, l.Password), nil
}

// FindLastLoginID 查询最新的登陆id
func FindLastLoginID() (int, error) {
	var id []int
	err := db.Table(logintable).Select("user_id").Order("user_id desc").Limit(1).Find(&FznewsLogin{}).Pluck("id", &id).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	if len(id) == 0 {
		return 0, nil
	}
	return id[0], nil

}

// encodePass 密码加密
func encodePass(id int, pass string) string {
	p := sha256.Sum256(append([]byte(fmt.Sprintf("%d", id)), []byte(pass)...))
	return string(Base58Encode(p[:]))
}

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode 将二进制数组编译成base58
func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}

	ReverseBytes(result)

	return result
}

// Base58Decode  decodes Base58-encoded data
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
