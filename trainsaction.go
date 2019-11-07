package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

//交易结构
type Transaction struct {
	//交易id
	Txid []byte
	//多个交易输入
	TxInput []TXInput
	//多个交易输出
	TxOutput []TXOutput
	//时间戳
	TimeStamp int64
}
type TXInput struct {
	//1、所引用的output的交易id
	TXID []byte
	//2、所引用的output的索引值
	Index int64
	//3、解锁脚本
	//ScriptSig string //包含私钥签名和公钥
	ScriptSig []byte //私钥签名
	PubKey    []byte
}
type TXOutput struct {
	//锁定脚本
	//LockScript string
	PubKeyHAsh []byte
	//转账金额
	Value float64
}

const reward = 12.5

//收款人给付款人地址，锁定的时候不是使用地址锁定的，而是使用公钥哈希锁定的
//提供一个生成output的方法
func NewTXOutput(value float64, address string) TXOutput {
	//计算公钥哈希
	pubKeyHash := getPubKetFromAddress(address)
	output := TXOutput{
		PubKeyHAsh: pubKeyHash,
		Value:      value,
	}
	return output
}

//挖矿交易
//没有引用的输入，只有输出，只有一个output
func NewCoinbaseTx(miner string, data string) *Transaction {
	inputs := []TXInput{{
		TXID:      nil,
		Index:     -1,
		ScriptSig: []byte(data),
		PubKey:    nil,
	}}
	output := NewTXOutput(reward, miner)
	outputs := []TXOutput{output}
	/*outputs := []TXOutput{{
		LockScript: miner,
		Value:      reward,
	}}*/
	tx := &Transaction{

		TxInput:   inputs,
		TxOutput:  outputs,
		TimeStamp: time.Now().Unix(),
	}
	tx.SetTxID()
	return tx
}

//设置当前交易的id，使用交易本身的哈希值作为自己交易id
func (tx *Transaction) SetTxID() {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(tx)
	if err != nil {
		fmt.Println("设置交易id失败, err:", err)
		return
	}
	hash := sha256.Sum256(buff.Bytes())
	tx.Txid = hash[:]
}

//普通交易
func NewTransaction(from, to string, amount float64, bc *BlockChain) (*Transaction, error) {
	//TODO
	//打开钱包
	wm := NewWalletMananger()
	if wm == nil {
		return nil, errors.New("打开钱包失败！")
	}
	//找到付款方对应的私钥和公钥
	w, ok := wm.Wallets[from]
	if !ok {
		return nil, fmt.Errorf("没有找到：`%s`对应的钱包", from)
	}
	//创建input的时候需要私钥签名和公钥
	//privKEy:=w.PrivKey//TODO

	pubKey := w.PubKey
	pubKeyHash := getPubKeyHashFromPubKey(pubKey)
	//1. 1. 找到付款人能够支配的合理的钱，返回金额和utxoinfo
	utxoinfos, value := bc.FindNeedUtxoInfo(pubKeyHash, amount)
	//2. 判断返回金额是否满足转账条件，如果不满足，创建交易失败。
	if value < amount {
		return nil, errors.New("付款人金额不足！")
	}
	//3. 拼接一个新的交易
	var inputs []TXInput
	var outputs []TXOutput
	//1. 拼装inputs
	for _, utxoinfo := range utxoinfos {
		input := TXInput{
			TXID:      utxoinfo.txid,
			Index:     utxoinfo.index,
			ScriptSig: nil, //钱包在交易创建的最后处理 TODO
			PubKey:    pubKey,
		}
		inputs = append(inputs, input)
	}
	//1. 遍历返回的utxonifo切片，逐个转成input结构
	//2. 拼装outputs
	//1. 拼装一个属于收款人的output
	output := NewTXOutput(amount, to)
	/*output := TXOutput{
		LockScript: to,
		Value:      amount,
	}*/
	outputs = append(outputs, output)
	//2. 判断一下是否需要找零，如果有，拼装一个属于付款方output
	if value > amount {
		//找零
		/*output1 := TXOutput{
			LockScript: from,
			Value:      value - amount,
		}*/
		output1 := NewTXOutput(value-amount, from)
		outputs = append(outputs, output1)
	}
	tx := Transaction{
		TxInput:   inputs,
		TxOutput:  outputs,
		TimeStamp: time.Now().Unix(),
	}
	//3. 设置交易id
	tx.SetTxID()
	//对当前交易进行签名
	bc.SignTransaction(w.PrivKey, &tx)
	//4. 返回
	return &tx, nil
}

//判断一个交易是否为挖矿交易
func (tx *Transaction) isCoinbaseTx() bool {
	input := tx.TxInput[0]
	if len(tx.TxInput) == 1 && input.Index == -1 {
		return true
	}
	return false
}

//创建当前交易的副本（裁剪） //Trim 修剪
func (tx *Transaction) TrimmedTransactionCopy() *Transaction {
	//将input的sig和pubkey字段设置成nil
	var inputs []TXInput
	var outputs []TXOutput
	//遍历input
	for _, input := range tx.TxInput {
		inputNew := TXInput{
			TXID:      input.TXID,
			Index:     input.Index,
			ScriptSig: nil,
			PubKey:    nil,
		}
		inputs = append(inputs, inputNew)

	}
	//遍历output
	copy(outputs, tx.TxOutput)
	txCopy := Transaction{
		Txid:      nil,
		TxInput:   inputs,
		TxOutput:  outputs,
		TimeStamp: tx.TimeStamp, //<< 不要使用当前时间，否则矿工校验时的数据一定会改变 }
	}
	return &txCopy
}

//具体签名函数
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey, prevTxs map[string]*Transaction) bool {
	fmt.Printf("开始具体签名动作：Sign ...\n")
	//所有的签名细节在此处实现 //TODO
	//1. 获取交易副本txCopy
	txCopy:=tx.TrimmedTransactionCopy()
	//遍历txCopy里面的input
	for i,input:=range txCopy.TxInput{
		prevtx:=prevTxs[string(input.TXID)]
		if privKey==nil{
			return false
		}
		//3. 使用这个input所引用的output来填充每一个input的pubKey字段
		output:=prevtx.TxOutput[input.Index]//<< == 不是i变量
		//input.PubKey = output.PubKeyHash <<== 这个input时副本，不要对它进行操作
		txCopy.TxInput[i].PubKey=output.PubKeyHAsh
		//对当前交易做哈希处理，得到需要签名的数据
		txCopy.SetTxID()
		hashData:=txCopy.Txid
		fmt.Printf("======》签名内容的哈希值：%x\n",hashData)
		//使用私钥进行签名:sig
		r,s,err:=ecdsa.Sign(rand.Reader,privKey,hashData[:])
		if err != nil {
			fmt.Println("ecdsa.Sign err:",err)
			return false
		}
		//将r，s拼接成一个字节流，生成当前的签名
		signature := append(r.Bytes(), s.Bytes()...)
		//6. 将签名赋值给原始交易的input.ScriptSig字段
		tx.TxInput[i].ScriptSig=signature
		//7. 将当前input的pubKey字段设置成nil
		txCopy.TxInput[i].PubKey=nil
		txCopy.Txid=nil//在SetTxId时，已经修改了，需要还原

	}
	fmt.Println("交易签名成功!")
	return true
}

//具体的验证函数
func (tx *Transaction) Verify(prevTxs map[string]*Transaction) bool {
	fmt.Printf("开始具体校验动作：Verify ...\n")
	//TODO
	//1. 生成一个交易的副本：txCopy
	txCopy:=tx.TrimmedTransactionCopy()
	//遍历交易当前交易副本，还原签名数据
	for i,input:=range txCopy.TxInput{
		//找到这个input对应的output，获取公钥哈希
	prevTx:=prevTxs[string(input.TXID)]
	if prevTxs==nil{
		return false
	}
	//对交易做哈希处理
	output:=prevTx.TxOutput[input.Index]
	txCopy.TxInput[i].PubKey=output.PubKeyHAsh
	//生成校验的数据
	txCopy.SetTxID()
	hashData:=txCopy.Txid//数据
	sigData:=tx.TxInput[i].ScriptSig//签名
	pubKey:=tx.TxInput[i].PubKey//公钥
		fmt.Printf(" ===> 校验时，还原的数据哈希值:%x\n", hashData)
		//还原签名，得到r,s
		var r, s big.Int
		r.SetBytes(sigData[:len(sigData)/2])
		s.SetBytes(sigData[len(sigData)/2:])
		//还原公钥，现在X，Y字节流
		var x, y big.Int
		x.SetBytes(pubKey[:len(pubKey)/2])
		y.SetBytes(pubKey[len(pubKey)/2:])
		pubKeyRaw := ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X: &x,
			Y: &y,
		}
		//3. 使用签名，公钥，数据，进行校验
		if !ecdsa.Verify(&pubKeyRaw, hashData[:], &r, &s) {
			//只要有一个input校验失败，则返回false
			fmt.Println("校验签名失败!")
			return false
		}
		//4. 清理数据，将相应的字段设置成nil
		txCopy.Txid = nil
		txCopy.TxInput[i].PubKey = nil
	}
	fmt.Println("当前交易数字签名校验成功!")
	return true
}
func (tx *Transaction) String() string {
	var lines []string
	lines=append(lines,fmt.Sprintf("---Transaction %x:",tx.Txid))
	for i,input:=range tx.TxInput{
		lines = append(lines, fmt.Sprintf(" Input %d:", i))
		lines = append(lines, fmt.Sprintf(" TXID: %x", input.TXID))
		lines = append(lines, fmt.Sprintf(" Out: %d", input.Index))
		lines = append(lines, fmt.Sprintf(" Signature: %x", input.ScriptSig))
		lines = append(lines, fmt.Sprintf(" PubKey: %x", input.PubKey))
	}
	for i, output := range tx.TxOutput{
		lines = append(lines, fmt.Sprintf(" Output %d:", i))
		lines = append(lines, fmt.Sprintf(" Value: %f", output.Value))
		lines = append(lines, fmt.Sprintf(" Script: %x", output.PubKeyHAsh))
	}
	return strings.Join(lines, "\n")
}
