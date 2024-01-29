package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IRepository[T IDatabaseObject] interface {
	Create(entity T, collectionName string) (*T, error)
	Count(filters interface{}, options *options.CountOptions, collectionName string) (int64, error)
	Get(filters interface{}, options *options.FindOneOptions, collectionName string) (*T, error)
	GetList(filters interface{}, options *options.FindOptions, collectionName string) ([]*T, error)
	Update(filters interface{}, entity *T, collectionName string) (*T, error)
	Remove(filters interface{}, collectionName string) error
}

type Repository[T IDatabaseObject] struct {
	context      *context.Context
	client       *mongo.Client
	DatabaseName string
}

func NewRepository[T IDatabaseObject](client *mongo.Client, databaseName string) *Repository[T] {
	return &Repository[T]{
		context:      new(context.Context),
		client:       client,
		DatabaseName: databaseName,
	}
}

func (r *Repository[T]) Create(entity T, collectionName string) (*T, error) {
	_, err := r.client.Database(r.DatabaseName).Collection(collectionName).InsertOne(*r.context, entity)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *Repository[T]) Get(filters interface{}, options *options.FindOneOptions, collectionName string) (*T, error) {
	var entity *T
	err := r.client.Database(r.DatabaseName).Collection(collectionName).FindOne(*r.context, filters, options).Decode(&entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Repository[T]) Count(filters interface{}, options *options.CountOptions, collectionName string) (int64, error) {
	count, err := r.client.Database(r.DatabaseName).Collection(collectionName).CountDocuments(*r.context, filters, options)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository[T]) GetList(filters interface{}, options *options.FindOptions, collectionName string) ([]*T, error) {
	var entities []*T
	cur, err := r.client.Database(r.DatabaseName).Collection(collectionName).Find(*r.context, filters, options)
	if err != nil {
		return nil, err
	}
	err = cur.All(context.TODO(), &entities)
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *Repository[T]) Update(filters interface{}, entity *T, collectionName string) (*T, error) {
	result, err := r.client.Database(r.DatabaseName).Collection(collectionName).ReplaceOne(*r.context, filters, entity)
	if err != nil {
		return nil, err
	}
	if result.ModifiedCount < 1 {
		return nil, fmt.Errorf("nothing has changed")
	}
	return entity, nil
}

func (r *Repository[T]) Remove(filters interface{}, collectionName string) error {
	result, err := r.client.Database(r.DatabaseName).Collection(collectionName).DeleteOne(*r.context, filters)
	if err != nil {
		return err
	}
	if result.DeletedCount < 1 {
		return fmt.Errorf("nothing has changed")
	}
	return nil
}
