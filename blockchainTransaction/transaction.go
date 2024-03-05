package blockchainTransaction

import (
	"github.com/btcsuite/btcd/repository"
	"github.com/btcsuite/btcd/wire"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BlockchainTransactionInput struct {
	TxIn wire.TxIn
	repository.DatabaseObject
	TransactionId string
	WitnessHash   string
	BlockHash     string
	Coin          string
}

type BlockchainTransaction struct {
	Id primitive.ObjectID `bson:"_id"`
	repository.DatabaseObject
	TransactionId          string
	WitnessHash            string
	Amount                 primitive.Decimal128
	ScriptClass            string
	Addresses              []string
	BlockHash              string
	RequiredSignatureCount int
	Coin                   string
}
