package main

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"log"
	"time"
)

func main() {

	client := rpc.New(rpc.MainNetBeta_RPC)
	wsClient, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}

	// 订阅token program日志事件
	// TOKEN_PROGRAM_ID Token Program address
	var TokenProgramId = solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	sub, err := wsClient.LogsSubscribeMentions(TokenProgramId, rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	fmt.Println("开始监听新的mint地址...")

	for {
		got, err := sub.Recv()
		if err != nil {
			log.Fatalf("接收日志事件失败: %v", err)
			continue
		}
		//异步处理日志事件
		go func() {
			if got.Value.Logs != nil {
				for _, logs := range got.Value.Logs {
					if logs == "Program log: Instruction: InitializeMint" {
						//spew.Dump(got)
						// 有可能查询失败，多次尝试
						time.Sleep(1 * time.Second)
						for i := 0; i < 10; i++ {
							txSig := solana.MustSignatureFromBase58(got.Value.Signature.String())
							// 获取Transaction详情
							out, err := client.GetTransaction(
								context.TODO(),
								txSig,
								&rpc.GetTransactionOpts{
									Encoding: solana.EncodingBase64,
								},
							)
							// 如果失败进行下一次查询
							if err != nil {
								continue
							}
							// 解析Transaction
							decodedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
							if err != nil {
								log.Fatalf("解析Transaction失败:%v", err)
							}
							spew.Dump(decodedTx)
							fmt.Println(decodedTx.Message.AccountKeys[1].String())
							// 成功，退出循环
							break
						}
					}
				}
			}
		}()

	}
}
