package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
)

type WalletManager struct {
	//1. 定义一个map来管理所有的钱包
	//2. key：地址
	// 3. value：wallet
	Wallets map[string]*Wallet
}
func NewWalletMananger() *WalletManager {
	var wm WalletManager
	wm.Wallets=make(map[string]*Wallet)
	//加载已经存在钱包，从wallet.dat //TODO


	err:=wm.loadFromFile()
	if err != nil {
		fmt.Println("loadFromFile err:",err)
		return nil
	}
	return &wm
}
func (wm *WalletManager) createWallet() (string, error) {
	//调用wallet结构的创建方法
	w:=NEwWallet()
	if w==nil{
		return "",errors.New("创建钱包失败！")
	}
	address:=w.getAddress()
	wm.Wallets[address]=w
	//填充自己的wallets结构
	// TODO ， 放入map结构，存储到磁盘
	err:=wm.saveToFile()
	if err != nil {
		fmt.Println("存储钱包失败,err:",err)
		return "",err
	}
	//返回地址
	return address,nil
}

const walletFileName  ="wallet.dat"
//1. 将wm结构写入到磁盘=》向map中添加数据
// 2. 使用gob对wm进行编码后写入文件
func (wm *WalletManager)saveToFile()error{
	var buff bytes.Buffer
	//对interface 数据进行注册
	gob.Register(elliptic.P256())
	encode:=gob.NewEncoder(&buff)
	err:=encode.Encode(wm)
	if err != nil {
		fmt.Println("saveToFile encode err44:",err)
		return err
	}
	//写入磁盘
	err=ioutil.WriteFile(walletFileName,buff.Bytes(),0600)
	if err != nil {
		fmt.Println("saveToFile writeFile err51:",err)
		return err
	}
return  nil
}
//加载钱包里的密钥对
func (wm *WalletManager)loadFromFile()error{
	//判断钱包文件是否存在
	if !isFileExist(walletFileName){
	//这是第一次执行进入的逻辑，不属于错误
		fmt.Println("钱包不存在，准备创建！")
		return nil
	}
	fmt.Println("钱包存在,准备读取....")
	data,err:=ioutil.ReadFile(walletFileName)
	if err != nil {
		fmt.Println("loadFromFile err",err)
		return err
	}
	//注册interface
	gob.Register(elliptic.P256())
	//解码
	decoder:=gob.NewDecoder(bytes.NewReader(data))
	err=decoder.Decode(wm)
	if err != nil {
		fmt.Println("loadFromFile DEcode err:",err)
		return err
	}
	return nil
}
func (wm *WalletManager)listAddresses()(addresses []string){
	for address,_:=range wm.Wallets{//1. 遍历map，获取所有的key值
		addresses=append(addresses,address)//2. 拼装成切片返回
	}
	//3. 将地址数组排序后返回
	//默认是升序
	sort.Strings(addresses)
	return
}