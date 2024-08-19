package main

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage interface {
	NewStorage(targetURI string, sourceURI string) storage
	copy(sourceCollection string, targetCollection string, sourceDatabase string, targetDatabase string)
	getTargetDatabases() ([]string, error)
	getSourceDatabases() ([]string, error)
	getTargetCollections(databaseName string) ([]collection, error)
	getSourceCollections(databaseName string) ([]collection, error)
}

type storage struct {
	targetURI string
	sourceURI string
}

// Initialize new storage instance
func newStorage(targetURI string, sourceURI string) storage {
	var s storage
	s.targetURI = targetURI
	s.sourceURI = sourceURI
	return s
}

// Copy data from given source database/collection to target database/collection deleting all data in target first.
func (s storage) copy(sourceCollection string, targetCollection string, sourceDatabase string, targetDatabase string) error {
	sOptions := options.Client().ApplyURI(s.sourceURI)
	sClient, err := mongo.Connect(context.Background(), sOptions)
	if err != nil {
		return err
	}
	defer sClient.Disconnect(context.Background())

	tOptions := options.Client().ApplyURI(s.targetURI)
	tClient, err := mongo.Connect(context.Background(), tOptions)
	if err != nil {
		return err
	}
	defer tClient.Disconnect(context.Background())

	// Get source and target collections
	sc := sClient.Database(sourceDatabase).Collection(sourceCollection)
	tc := tClient.Database(targetDatabase).Collection(targetCollection)

	// Check there are documents to move
	count, err := s.getRecordCount(sClient, sourceDatabase, sourceCollection)
	if count == 0 {
		return errors.New("no records in source collection to copy")
	} else if err != nil {
		return err
	}

	// Find documents in the source collection
	cursor, err := sc.Find(context.Background(), bson.D{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	// Delete all documents in target
	tc.DeleteMany(context.Background(), bson.D{})

	// Iterate through documents and insert into target collection
	for cursor.Next(context.Background()) {
		var doc interface{}
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		var opts = options.InsertOneOptions{}
		_, err := tc.InsertOne(context.Background(), doc, &opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s storage) getRecordCount(client *mongo.Client, databaseName string, collectionName string) (int64, error) {
	// Get source and target collections
	sc := client.Database(databaseName).Collection(collectionName)

	// Check there are documents to move
	count, err := sc.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Get collections from target database.
func (s storage) getTargetCollections(databaseName string) ([]collection, error) {
	options := options.Client().ApplyURI(s.sourceURI)
	client, err := mongo.Connect(context.Background(), options)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	db := client.Database(databaseName)

	// Retrieve collection names
	c, err := db.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var collections []collection

	for _, name := range c {
		count, err := s.getRecordCount(client, databaseName, name)
		if err != nil {
			return collections, err
		}

		var collection collection
		collection.count = count
		collection.name = name
		collections = append(collections, collection)

	}

	return collections, nil
}

// Get collections from source database.
func (s storage) getSourceCollections(databaseName string) ([]collection, error) {
	options := options.Client().ApplyURI(s.sourceURI)
	client, err := mongo.Connect(context.Background(), options)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	db := client.Database(databaseName)

	// Retrieve collection names
	c, err := db.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var collections []collection

	for _, name := range c {
		count, err := s.getRecordCount(client, databaseName, name)
		if err != nil {
			return collections, err
		}

		var collection collection
		collection.count = count
		collection.name = name
		collections = append(collections, collection)
	}

	return collections, nil
}

// Get all databases from target server provided in config.
func (s storage) getTargetDatabases() ([]string, error) {
	options := options.Client().ApplyURI(s.targetURI)
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

// Get all databases from source server provided in config.
func (s storage) getSourceDatabases() ([]string, error) {
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
