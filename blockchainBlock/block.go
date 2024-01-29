package blockchainBlock

import (
	"github.com/btcsuite/btcd/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type BlockchainBlock struct {
	Id primitive.ObjectID `bson:"_id"`
	repository.DatabaseObject
	Version           int32
	Hash              string
	PreviousBlockHash string
	MerkleRoot        string
	Timestamp         time.Time
	Bits              uint32
	Nonce             uint32
	Height            uint32
	Coin              string
}
