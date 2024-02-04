// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/blockchainBlock"
	"github.com/btcsuite/btcd/blockchainTransaction"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/repository"
	"github.com/btcsuite/btcd/shared"
	"github.com/btcsuite/btcd/txscript"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/database"
)

// maybeAcceptBlock potentially accepts a block into the block chain and, if
// accepted, returns whether or not it is on the main chain.  It performs
// several validation checks which depend on its position within the block chain
// before adding it.  The block is expected to have already gone through
// ProcessBlock before calling this function with it.
//
// The flags are also passed to checkBlockContext and connectBestChain.  See
// their documentation for how the flags modify their behavior.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) maybeAcceptBlock(block *btcutil.Block, flags BehaviorFlags) (bool, error) {
	// The height of this block is one more than the referenced previous
	// block.
	prevHash := &block.MsgBlock().Header.PrevBlock
	prevNode := b.index.LookupNode(prevHash)
	if prevNode == nil {
		str := fmt.Sprintf("previous block %s is unknown", prevHash)
		return false, ruleError(ErrPreviousBlockUnknown, str)
	} else if b.index.NodeStatus(prevNode).KnownInvalid() {
		str := fmt.Sprintf("previous block %s is known to be invalid", prevHash)
		return false, ruleError(ErrInvalidAncestorBlock, str)
	}

	blockHeight := prevNode.height + 1
	block.SetHeight(blockHeight)

	// The block must pass all of the validation rules which depend on the
	// position of the block within the block chain.
	err := b.checkBlockContext(block, prevNode, flags)
	if err != nil {
		return false, err
	}

	// Insert the block into the database if it's not already there.  Even
	// though it is possible the block will ultimately fail to connect, it
	// has already passed all proof-of-work and validity tests which means
	// it would be prohibitively expensive for an attacker to fill up the
	// disk with a bunch of blocks that fail to connect.  This is necessary
	// since it allows block download to be decoupled from the much more
	// expensive connection logic.  It also has some other nice properties
	// such as making blocks that never become part of the main chain or
	// blocks that fail to connect available for further analysis.
	err = b.db.Update(func(dbTx database.Tx) error {
		return dbStoreBlock(dbTx, block)
	})
	if err != nil {
		return false, err
	}

	// Create a new block node for the block and add it to the node index. Even
	// if the block ultimately gets connected to the main chain, it starts out
	// on a side chain.
	blockHeader := &block.MsgBlock().Header
	newNode := newBlockNode(blockHeader, prevNode)
	newNode.status = statusDataStored
	br := repository.NewRepository[*blockchainBlock.BlockchainBlock](b.dbClient, shared.DatabaseName)
	tr := repository.NewRepository[*blockchainTransaction.BlockchainTransaction](b.dbClient, shared.DatabaseName)
	dbObj, err := br.Get(bson.D{{"height", blockHeight}, {"coin", shared.Bitcoin_Coin_Name}}, nil, repository.BlockCollectionName)
	if dbObj == nil {
		databaseBlock := &blockchainBlock.BlockchainBlock{
			Id: primitive.NewObjectID(),
			DatabaseObject: repository.DatabaseObject{
				UpdatedAt: time.Now().UTC(),
				CreatedAt: time.Now().UTC(),
				IsActive:  true,
			},
			Version:           blockHeader.Version,
			Hash:              blockHeader.BlockHash().String(),
			PreviousBlockHash: blockHeader.PrevBlock.String(),
			MerkleRoot:        blockHeader.MerkleRoot.String(),
			Timestamp:         blockHeader.Timestamp,
			Bits:              blockHeader.Bits,
			Nonce:             blockHeader.Nonce,
			Height:            uint32(blockHeight),
			Coin:              shared.Bitcoin_Coin_Name,
		}
		insertedBlock, err := br.Create(databaseBlock, repository.BlockCollectionName)
		if err != nil {
			fmt.Println(err)
		}
		databaseBlock = *insertedBlock
		for _, transaction := range block.Transactions() {
			for _, out := range transaction.MsgTx().TxOut {
				script, err := hex.DecodeString(fmt.Sprintf("%x", out.PkScript))
				if err != nil {
					fmt.Println(err)
				}
				// Extract and print details from the script.
				scriptClass, addresses, reqSigs, err := txscript.ExtractPkScriptAddrs(
					script, &chaincfg.MainNetParams)
				bitcoinAddresses := []string{}
				for _, address := range addresses {
					bitcoinAddresses = append(bitcoinAddresses, address.String())
				}
				stringAmount := strconv.FormatInt(out.Value, 10)
				transactionAmount, err := primitive.ParseDecimal128(stringAmount)
				if err != nil {

				}
				databaseTransaction := &blockchainTransaction.BlockchainTransaction{
					Id: primitive.NewObjectID(),
					DatabaseObject: repository.DatabaseObject{
						UpdatedAt: time.Now().UTC(),
						CreatedAt: time.Now().UTC(),
						IsActive:  true,
					},
					TransactionId:          transaction.Hash().String(),
					WitnessHash:            transaction.WitnessHash().String(),
					Amount:                 transactionAmount,
					ScriptClass:            scriptClass.String(),
					BlockHash:              blockHeader.BlockHash().String(),
					Addresses:              bitcoinAddresses,
					RequiredSignatureCount: reqSigs,
					Coin:                   shared.Bitcoin_Coin_Name,
				}
				insertedTransaction, err := tr.Create(databaseTransaction, repository.TransactionCollectionName)
				if err != nil {
					fmt.Println(err)
				}
				databaseTransaction = *insertedTransaction
			}
		}
	}
	b.index.AddNode(newNode)
	err = b.index.flushToDB()
	if err != nil {
		return false, err
	}

	// Connect the passed block to the chain while respecting proper chain
	// selection according to the chain with the most proof of work.  This
	// also handles validation of the transaction scripts.
	isMainChain, err := b.connectBestChain(newNode, block, flags)
	if err != nil {
		return false, err
	}

	// Notify the caller that the new block was accepted into the block
	// chain.  The caller would typically want to react by relaying the
	// inventory to other peers.
	b.chainLock.Unlock()
	b.sendNotification(NTBlockAccepted, block)
	b.chainLock.Lock()

	return isMainChain, nil
}
