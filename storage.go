package main

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type storage struct {
	targetURI string
	sourceURI string
}

func NewStorage(targetURI string, sourceURI string) storage {
	var s storage
	s.targetURI = targetURI
	s.sourceURI = sourceURI
	return s
}

func (s storage) copy(collection string, database string) error {
	sOptions := options.Client().ApplyURI(s.targetURI)
	sClient, err := mongo.Connect(context.Background(), sOptions)
	if err != nil {
		return err
	}
	defer sClient.Disconnect(context.Background())

	tOptions := options.Client().ApplyURI(s.sourceURI)
	tClient, err := mongo.Connect(context.Background(), tOptions)
	if err != nil {
		return err
	}
	defer tClient.Disconnect(context.Background())

	// Get source and target collections
	sourceCollection := sClient.Database(database).Collection(collection)
	targetCollection := tClient.Database("local2").Collection(collection)

	// Check there are documents to move
	count, err := sourceCollection.CountDocuments(context.Background(), bson.D{})
	if count == 0 {
		return errors.New("no records in source collection to copy")
	} else if err != nil {
		return err
	}

	// Find documents in the source collection
	cursor, err := sourceCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	// Delete all documents in target
	targetCollection.DeleteMany(context.Background(), bson.D{})

	// Iterate through documents and insert into target collection
	for cursor.Next(context.Background()) {
		var doc interface{}
		if err := cursor.Decode(&doc); err != nil {
			if err != nil {
				return err
			}
		}

		var opts = options.InsertOneOptions{}
		_, err := targetCollection.InsertOne(context.Background(), doc, &opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s storage) getCollections(databaseName string) ([]string, error) {
	options := options.Client().ApplyURI(s.sourceURI)
	client, err := mongo.Connect(context.Background(), options)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	db := client.Database(databaseName)

	// Retrieve collection names
	result, err := db.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s storage) getDatabases() ([]string, error) {
	options := options.Client().ApplyURI(s.sourceURI)
	client, err := mongo.Connect(context.Background(), options)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	result, err := client.ListDatabaseNames(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	return result, nil
}
