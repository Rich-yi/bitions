package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	PrivKey *ecdsa.PrivateKey
	PubKey  []byte //这不是原生公钥，而是x,y字节流拼成的
}

//创建一个密钥对
func NEwWallet() *Wallet {
	//私钥
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("创建密钥对失败，err", err)
		return nil
	}
	//公钥
	pubKeyRaw := privKey.PublicKey
	x := pubKeyRaw.X
	y := pubKeyRaw.Y
	pubKey := append(x.Bytes(), y.Bytes()...)
	return &Wallet{
		PrivKey: privKey,
		PubKey:  pubKey,
	}
}
func (w *Wallet) getAddress() string {
//通过公钥，获取公钥哈希
	pubKeyHash:=getPubKeyHashFromPubKey(w.PubKey)
	//二、在前面添加1个字节的版本号
	payload := append([]byte{byte(00)}, pubKeyHash...)
	//三、做两次哈希运算，截取前四个字节，作为checksum，
	checksum:=checkSum(payload)
	//四、拼接25字节数据
	payload = append(payload, checksum...)
	//五、base58处理，得到地址
	address := base58.Encode(payload)
	return address

}

//通过公钥获取公钥哈希
func getPubKeyHashFromPubKey(pubKey []byte) []byte {
	//第一次哈希
	firstHash := sha256.Sum256(pubKey)
	//第二次哈希
	hasher := ripemd160.New()
	hasher.Write(firstHash[:])
	pubKeyHash := hasher.Sum(nil)
	return pubKeyHash
}

//通过地址获取公钥哈希
//base58解码。得到25字节数据
func getPubKetFromAddress(address string) []byte {
	decodeInfo := base58.Decode(address)
	if len(decodeInfo) != 25 {
		fmt.Println("地址长度无效, 应该为25字节，当前字节为:", len(decodeInfo))
		return nil
	}
	//截取中间20字节
	pubkeyHash := decodeInfo[1:21]
	return pubkeyHash
}
func checkSum(payload []byte)[]byte{
	f1 := sha256.Sum256(payload)
	second := sha256.Sum256(f1[:])
	checksum := second[:4] //左闭右开
	return checksum
}

const payloadLen =25
const checkSumLen  = 4
func isValidAddress(address string) bool {
	//1. 对传入的地址解密=》得到25字节数据
	decodeInfo:=base58.Decode(address)
	if len(decodeInfo) != payloadLen {
		return false
	}
	//2. 截取出前21byte，计算checksum，得到checksum1
	checksum1:=checkSum(decodeInfo[:payloadLen-checkSumLen])
	//3. 截取后4byte，得到checksum2
	checksum2 := decodeInfo[payloadLen-checkSumLen:]
	fmt.Printf("checksum1: %x\n", checksum1)
	//4. 比较checksum与checksum2，
	fmt.Printf("checksum2: %x\n", checksum2)
	return bytes.Equal(checksum1,checksum2)
	//1. 相同：地址有效 //2. 不同：地址无效
}
