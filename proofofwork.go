package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type ProofOfWork struct {
	block *Block//区块数据block
	target *big.Int//难度值target

}
func NewProofOfWork(block *Block)*ProofOfWork{
	pow:=ProofOfWork{
		block:  block,
	}
	targetStr:="0010000000000000000000000000000000000000000000000000000000000000"
	bigTmp:=big.Int{}
	bigTmp.SetString(targetStr,16)
	pow.target=&bigTmp
	return &pow

}
func (pow *ProofOfWork)prepareData(nonce int64)[]byte{
	b:=pow.block
	tmp:=[][]byte{
		[]byte(b.Version),
		b.PrevHash,
		b.MerkleRoot,

		dig2byte(b.TimeStamp),
		dig2byte(b.Bits),
		dig2byte(nonce),
		//b.Data,//TODO
	}
	data:=bytes.Join(tmp,[]byte(""))
	return data
}
//核心函数，不断改变Nonce，求出满足条件的哈希值
func (pow *ProofOfWork) Run() (int64, []byte) {
	var nonce int64
	var hash [32]byte
	for {
		fmt.Printf("挖矿中:%x\r", hash)
		//拼好的数据
		data := pow.prepareData(nonce)
		//计算哈希值
		hash = sha256.Sum256(data)
		//转换成big.int
		tmpInt := big.Int{}
		tmpInt.SetBytes(hash[:])
		//比较 // -1 if x < y
		// 0 if x == y
		// +1 if x > y
		// func (x *Int) Cmp(y *Int) (r int)
		if tmpInt.Cmp(pow.target) == -1 {
			fmt.Printf("挖矿成功，当前哈希值为:%x, nonce: %d\n", hash, nonce)
			break
		} else {
			nonce++
		}

	}
	return nonce,hash[:]
}