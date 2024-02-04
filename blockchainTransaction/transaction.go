package blockchainTransaction

import (
	"github.com/btcsuite/btcd/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
