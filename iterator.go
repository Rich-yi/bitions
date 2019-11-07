package main

import (
	"github.com/boltdb/bolt"
	"log"
)

type Iterator struct {
	db *bolt.DB
	currHash []byte//当前区块的哈希，会不断变化
}
func NewIterator(bc *BlockChain)*Iterator{
	return &Iterator{
		db:       bc.db,
		currHash: bc.tail,
	}
}
//会多次调用这个方法，每次调用时
//1. 会返回当前位置的区块
//2. 游标currHash 会指向前一个区块
func (it *Iterator)Next()*Block{
var block *Block
it.db.View(func(tx *bolt.Tx) error {
	//获取bucket
	b:=tx.Bucket([]byte(blockBucket))
	if b==nil{
		log.Fatal("Iterator Next时，bucket不应为空!")
	}
	//获取区块
	blockInfo:=b.Get(it.currHash)
	block=Deserialize(blockInfo)
	//游标左移
	it.currHash=block.PrevHash

	return nil
})
	return block
}