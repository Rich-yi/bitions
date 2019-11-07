package main

import "fmt"

func (cli *CLI) addBlock(data string) {
	//cli.bc.AddBlock(data)//TODO
}
func (cli *CLI) printBlock() {
	it := NewIterator(cli.bc)
	for {
		block := it.Next()
		fmt.Println("+++++++")
		fmt.Printf("version:%s\n", block.Version)
		fmt.Printf("prevHash: %x\n", block.PrevHash)
		fmt.Printf("hash: %x\n", block.Hash)
		fmt.Printf("merkleRoot:%x\n", block.MerkleRoot)
		fmt.Printf("timeStamp:%d\n", block.TimeStamp)
		fmt.Printf("bits:%d\n", block.Bits)
		fmt.Printf("nonce:%d\n", block.Nonce)
		fmt.Printf("data: %s\n", block.Transactions[0].TxInput[0].ScriptSig)
		//判断是否已经是创世区块
		if block.PrevHash == nil {
			break
		}
	}
	fmt.Println("区块链遍历结束!")
}
func (cli *CLI) getBalance(address string) {
	if !isValidAddress(address){
		fmt.Println("无效的地址：",address)
		return
	}
	//utxos := cli.bc.FindMyUtxo(address)
	//通过地址获取公钥哈希
	pubKeyHash:=getPubKetFromAddress(address)
	utxos:=cli.bc.FindMyUtxo(pubKeyHash)
	var total float64
	for _, utxo := range utxos {
		//total+=utxo.Value
		total += utxo.output.Value
	}
	fmt.Printf("`%s`的比特币余额为：%f\n", address, total)
}
func (cli *CLI) send(from, to string, amount float64, miner, data string) {
	fmt.Printf("'%s'向'%s转账:'%f', miner:%s, data:%s\n", from, to, amount, miner, data)
	//输入数据的有效性会进行校验 //TODO
	if !isValidAddress(from){
		fmt.Println("无效的from地址：",from)
		return
	}
	if !isValidAddress(to){
		fmt.Println("无效的to地址：",to)
		return
	}
	if !isValidAddress(miner){
		fmt.Println("无效的miner地址：",miner)
		return
	}
	//创建挖矿交易
	coninbaseTx := NewCoinbaseTx(miner, data)
	txs := []*Transaction{coninbaseTx}
	//一个区块只添加一笔有效的普通交易
	tx, err := NewTransaction(from, to, amount, cli.bc)
	if err != nil {
		fmt.Println("err:", err)
	} else {
		fmt.Printf("发现有效的交易，准备添加到区块, txid:%x\n", tx.Txid)
		txs = append(txs, tx)
	}
	//创建区块，添加到区块链
	cli.bc.AddBlock(txs)
}
func (cli *CLI) createWallet() {
	wm:=NewWalletMananger()
	if wm==nil{
		fmt.Println("打开钱包失败")
		return
	}
	address, err := wm.createWallet()
	if err != nil {
		fmt.Println("创建钱包失败,err",err)
		return
	}
	fmt.Println("创建新地址成功",address)
}
func (cli *CLI)listAddress(){
	wm:=NewWalletMananger()
	if  wm== nil {
		fmt.Println("打开钱包失败！")
		return
	}
	addresses:=wm.listAddresses()
	for _,address:=range addresses{
		fmt.Println("address：",address)
	}
}
func (cli *CLI) printtx() {
	it:=NewIterator(cli.bc)
	for  {
		block := it.Next()
		fmt.Println("++++++++++++++++++++++")
		for _, tx := range block.Transactions {
			//自定义了Transaction的String()
			fmt.Println(tx)
		}
		if len(block.PrevHash) == 0 { break }
	}
}