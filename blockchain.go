package main

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

type BlockChain struct {
	//blocks []*Block
	db   *bolt.DB //句柄，数据库对象handler
	tail []byte   //存储最后一个区块哈希值
}

const genesisInfo = "多喝烫水"
const blockChainFileName = "blockchain.db"
const blockBucket = "blockBucket"
const lastBlockHashKey = "lastBlockHashKey"

//创建blockchain 添加创世区块
func NewBlockChain() *BlockChain {
	var lastHash []byte
	db, err := bolt.Open(blockChainFileName, 0600, nil)
	if err != nil {
		fmt.Println("创建区块链失败, err:", err)
		return nil
	}
	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		if b == nil {
			b, err = tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				fmt.Println("创建bucket失败, err:", err)
				return err
			}
			//写入创世块
			//创建一个挖矿交易，里面写入创世语
			coinbaseTx := NewCoinbaseTx("1QJUkhLJqTgWnpgGhh64dYgPiig7QgwxfK", genesisInfo)
			genesisBlock := CreateBlock([]*Transaction{coinbaseTx}, nil)
			_ = b.Put(genesisBlock.Hash, genesisBlock.Serialize() /*区块转换成字节流 */)
			_ = b.Put([]byte(lastBlockHashKey), genesisBlock.Hash)
			hash := b.Get([]byte(lastBlockHashKey))
			fmt.Printf("lastHash:%x\n", hash)
			//获取数据库中block序列化之后的数据
			blockInfo := b.Get(genesisBlock.Hash)
			block := Deserialize(blockInfo)
			fmt.Printf("block from db :%x\n", block)
			lastHash = genesisBlock.Hash
		} else { //bucket已经存在, 直接读取最后一区块的哈希值
			lastHash = b.Get([]byte(lastBlockHashKey))
		}
		return nil
	})

	return &BlockChain{
		db:   db,
		tail: lastHash,
	}
}

//添加区块方法
func (bc *BlockChain) AddBlock(transaction []*Transaction) {
	fmt.Println("AddBlock called!")
	//所有校验通过的交易集合，最终打包到区块 <<========
	var validTxs []*Transaction
	//对每一条交易进行校验
	for _, tx := range transaction {
		if bc.VerifyTransaction(tx) {
			validTxs = append(validTxs, tx)
		} else {
			fmt.Printf("发现签名校验失败的交易:%x\n", tx.Txid)
		}
	}

	//最后一个区域的哈希值
	lastHash := bc.tail
	//1、创建新的区块
	newBlock := CreateBlock(validTxs, lastHash)
	bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		if b == nil {
			log.Fatal("AddBlock是，bucket不应为空！")
		}
		//写入区块
		_ = b.Put(newBlock.Hash, newBlock.Serialize())
		_ = b.Put([]byte(lastBlockHashKey), newBlock.Hash)
		//更新tail的值

		bc.tail = newBlock.Hash

		return nil
	})
	//lastBlock := b.blocks[len(b.blocks)-1]
	//prevHash := lastBlock.hash
	//newBlock := CreateBlock(data, prevHash) //1. 创建新的区块
	//b.blocks = append(b.blocks, newBlock)   //2. 添加到bc的blocks
}

//定义一个结构体，同时包含output和它的信息位置
type UtxoInfo struct {
	output TXOutput
	index  int64
	txid   []byte
}

//遍历账本，查询指定地址所有的utxo
func (bc *BlockChain) FindMyUtxo(pubKeyHash []byte) []UtxoInfo {
	fmt.Printf("FindMyUtxo called ,pubKeyHash:%x\n", pubKeyHash)
	//var outputs []TXOutput
	var utxoinfos []UtxoInfo
	//1. 遍历区块
	//定义一个map，用于存储已经消耗过的output
	spentOutput := make(map[string][]int64)
	it := NewIterator(bc)
	for {
		block := it.Next()
		///2. 遍历交易
		for _, tx := range block.Transactions {
			//遍历output
		LABEL1:
			for outputIndex, output := range tx.TxOutput {
				//判断当前的output是否是目标地址锁定的
				//if output.LockScript == address {
				if bytes.Equal(output.PubKeyHAsh, pubKeyHash) {
					//查看当前交易是否已经存在与spentoutput容器中
					currTxId := string(tx.Txid)
					indexArr := spentOutput[currTxId]
					if len(indexArr) != 0 {
						//说明容器存在当前交易的output
						for _, spenyIndex /*0,1*/ := range indexArr {
							if outputIndex == int(spenyIndex) {
								fmt.Println("当前的output已经被使用过了,无需统计，index：", outputIndex)
								continue LABEL1
							}
						}

					}
					fmt.Printf("找到了属于'%x'的output, index:%d, value:%f\n", pubKeyHash, outputIndex, output.Value)
					utxinfo := UtxoInfo{
						output: output,
						index:  int64(outputIndex),
						txid:   tx.Txid,
					}
					utxoinfos = append(utxoinfos, utxinfo)
					// outputs = append(outputs, output)
				}
			}
			//遍历inputs， 得到一个map //TODO
			if !tx.isCoinbaseTx() {
				//遍历input是，得到一个map
				for _, input := range tx.TxInput {
					pubKeyHash1 := getPubKeyHashFromPubKey(input.PubKey)
					if bytes.Equal(pubKeyHash1, pubKeyHash) {
						//if input.ScriptSig == address {
						spentKey := string(input.TXID)

						spentOutput[spentKey] = append(spentOutput[spentKey], input.Index)
					}
				}
			}
		}
		if block.PrevHash == nil {
			break
		}
	}
	return utxoinfos
}
func (bc *BlockChain) FindNeedUtxoInfo(pubKeyHash []byte, amount float64) ([]UtxoInfo, float64) {
	fmt.Printf("FindNeedUtxoInfo called,adress:%x.amount:%f\n", pubKeyHash, amount)
	///1. 遍历账本，找到所有address（付款人）的utxo集合
	utxoinfos := bc.FindMyUtxo(pubKeyHash)
	//返还的utxoinfo里面包含金额
	var retValue float64
	var retUtxoInfo []UtxoInfo
	//2. 筛选出满足条件的数量即可，不要全部返还
	for _, utxoinfo := range utxoinfos {
		retUtxoInfo = append(retUtxoInfo, utxoinfo)
		retValue += utxoinfo.output.Value
		if retValue >= amount {
			//满足转账需求，直接返回
			break
		}
	}
	return retUtxoInfo, retValue
}
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	fmt.Printf("开始校验：VerifyTransaction called!\n")
	if tx.isCoinbaseTx() {
		fmt.Println("发现挖矿交易，不需要校验！")
		return true
	}
	prevTxs := make(map[string]*Transaction)
	//TODO
	for _, input := range tx.TxInput {
		tx := bc.FindTransactionByTxid(input.TXID)
		if tx == nil {
			fmt.Println("没有找到交易, txid:", input.TXID)
			return false
		}
		//将找到的交易放到集合中
		fmt.Printf("找到校验时所引用的交易,txid: %x\n", tx.Txid)
		prevTxs[string(input.TXID)] = tx
	}

	return tx.Verify(prevTxs)
}

//根据txid找到交易本身
func (bc *BlockChain) FindTransactionByTxid(txid []byte) *Transaction {
	it := NewIterator(bc)
	for {
		block := it.Next()
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.Txid, txid) {

				return tx
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return nil
}

//签名相关
func (bc *BlockChain) SignTransaction(priKey *ecdsa.PrivateKey, tx *Transaction) bool {
	fmt.Printf("开始签名：SignTransaction called!\n")
	if tx.isCoinbaseTx() {
		fmt.Println("发现挖矿交易，不需要签名!")
		return true
	}
	//1. 查到tx所引用的交易的集合 <<<======
	//key: txid
	// value:tx本身
	prevTxs := make(map[string]*Transaction)
	//遍历当前交易的input，通过txid找到每个input的tx，赋值给map
	for _, input := range tx.TxInput {
		tx := bc.FindTransactionByTxid(input.TXID)
		if tx == nil {
			fmt.Println("没有找到交易, txid:", input.TXID)
			return false
		}
		//将找到的交易放到集合中
		fmt.Printf("找到签名时所引用的交易,txid: %x\n", tx.Txid)
		prevTxs[string(input.TXID)] = tx
	}
	return tx.Sign(priKey, prevTxs)
}
