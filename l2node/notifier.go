package l2node

import (
	"fmt"
	"time"
)

type BlockData struct {
	Height   int64
	Txs      [][]byte
	L2Config []byte
	ZKConfig []byte
	Root     []byte
}

type Notifier struct {
	txsAvailable chan struct{}
	blockData    *BlockData
	l2Node       L2Node
}

func NewNotifier(l2Node L2Node) *Notifier {
	return &Notifier{
		txsAvailable: make(chan struct{}, 1),
		blockData:    nil,
		l2Node:       l2Node,
	}
}

func (n *Notifier) TxsAvailable() <-chan struct{} {
	return n.txsAvailable
}

func (n *Notifier) RequestBlockData(height int64, createEmptyBlocksInterval time.Duration) {
	if createEmptyBlocksInterval == 0 {
		go func() {
			for {
				txs, configs, err := n.l2Node.RequestBlockData(height)
				if err != nil {
					fmt.Println("ERROR:", err.Error())
					return
				}
				if len(txs) > 0 {
					n.blockData = &BlockData{
						Height:   height,
						Txs:      txs,
						L2Config: configs.L2Config,
						ZKConfig: configs.ZKConfig,
						Root:     configs.Root,
					}
					n.txsAvailable <- struct{}{}
					return
				}
				time.Sleep(500 * time.Millisecond)
			}
		}()
	} else {
		timeout := time.After(createEmptyBlocksInterval)
		go func() {
			for {
				select {
				case <-timeout:
					return
				default:
					txs, configs, err := n.l2Node.RequestBlockData(height)
					if err != nil {
						fmt.Println("ERROR:", err.Error())
						return
					}
					if len(txs) > 0 {
						n.blockData = &BlockData{
							Height:   height,
							Txs:      txs,
							L2Config: configs.L2Config,
							ZKConfig: configs.ZKConfig,
							Root:     configs.Root,
						}
						n.txsAvailable <- struct{}{}
						return
					}
				}
				time.Sleep(500 * time.Millisecond)
			}
		}()
	}
}

func (n *Notifier) GetBlockData() *BlockData {
	return n.blockData
}

func (n *Notifier) CleanBlockData() {
	n.blockData = nil
}
