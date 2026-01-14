package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// CalculateDataHash 计算数据哈希(用于上链)
// dataHash = SHA256(bizId + amount + salt)
func CalculateDataHash(bizId, amount, salt string) string {
	data := fmt.Sprintf("%s%s%s", bizId, amount, salt)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CalculateBizIdHash 计算业务流水号哈希
// 用于将字符串流水号转换为bytes32格式
func CalculateBizIdHash(bizId string) string {
	hash := sha256.Sum256([]byte(bizId))
	return "0x" + hex.EncodeToString(hash[:])
}

// VerifyDataHash 验证数据哈希
func VerifyDataHash(bizId, amount, salt, expectedHash string) bool {
	calculatedHash := CalculateDataHash(bizId, amount, salt)
	return calculatedHash == expectedHash
}
