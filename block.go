package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"
)

type Block struct {
	//版本号
	Version    string
	PrevHash   []byte //前区块哈希
	MerkleRoot []byte //merkle根, 根据当前区块的交易数据计算出来的
	Hash       []byte //当前区块，比特币中比特币中没有当前区块这个字段，为了方便处理
	TimeStamp  int64  //时间戳
	Nonce      int64  //难度值，系统提供的
	Bits       int64  //随机数
	//Data       []byte //区块体，交易数据
	Transactions []*Transaction
}

func CreateBlock(Transactions []*Transaction, prevHash []byte) *Block {
	block := &Block{
		Version:  "0",
		PrevHash: prevHash,
		//MerkleRoot:nil,
		TimeStamp:  time.Now().Unix(),
		Nonce:       0,
		Bits:       0,
		Transactions: Transactions,
	}
	//block.setHash()
	//将梅克尔根暂且省略
	block.HashTransaction()
	pow:=NewProofOfWork(block)
	nonce,hash:=pow.Run()
	block.Nonce=nonce
	block.Hash=hash
	return block
}
/*func (b *Block) setHash() {
	//data1 := append([]byte(b.version), b.prevHash...)
	//data1 = append(data1, b.data...)
	tmp:=[][]byte{
		[]byte(b.Version),
		b.PrevHash,
		b.MerkleRoot,
		//[]byte(string(b.timeStamp)),
		//[]byte(string(b.bits)),
		//[]byte(string(b.nonce)),
		dig2byte(b.TimeStamp),
		dig2byte(b.Bits),
		dig2byte(b.Nonce),
		//b.Data,

	}
	data1:=bytes.Join(tmp,[]byte(""))
	hash := sha256.Sum256(data1)
	b.Hash = hash[:]

}*/
//序列化区块
//结构=》编码=》字节流
func (b *Block)Serialize()[]byte  {
	//创建编码器
	var buff bytes.Buffer
	encode:=gob.NewEncoder(&buff)
	//编码
	err:=encode.Encode(b)
	if err != nil {
		fmt.Println("区块编码失败, err:", err)
		return nil
	}
	return buff.Bytes()
}
//反序列化区块
//字节流=》解码=》结构
func Deserialize(data []byte)*Block  {
	var block Block
	//创建解码器
	decoder:=gob.NewDecoder(bytes.NewReader(data))
	//解码
	err:=decoder.Decode(&block)
	if err != nil {
		fmt.Println("区块解析字节流失败，err:",err)
		return nil
	}
	return &block
}
//模拟计算梅克尔根
func (b *Block) HashTransaction() {
	var info []byte
	for _,tx:=range b.Transactions{
		info=append(info,tx.Txid...)
	}
	hash:=sha256.Sum256(info)
	b.MerkleRoot=hash[:]
}