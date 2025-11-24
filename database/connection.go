package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" 
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Variabel Global Exported
// Gunakan variabel ini di Repository/Service Anda nanti
var (
	PgDB        *sql.DB         // Koneksi PostgreSQL
	MongoClient *mongo.Client   // Koneksi Client Mongo
	MongoDB     *mongo.Database // Koneksi Database Mongo Spesifik
)

func ConnectDB() {
	connectPostgres()
	connectMongo()
	log.Println("Success: All database connections established")
}

func connectPostgres() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), 
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	PgDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(" Error opening Postgres connection:", err)
	}

	// Cek ping untuk memastikan koneksi hidup
	if err = PgDB.Ping(); err != nil {
		log.Fatal(" Error connecting to Postgres:", err)
	}

	PgDB.SetMaxOpenConns(25)
	PgDB.SetMaxIdleConns(25)
	PgDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Connected to PostgreSQL successfully")
}

func connectMongo() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable is not set")
	}

	dbName := os.Getenv("MONGO_DB_NAME") // Pisahkan nama DB Mongo dan Postgres
	if dbName == "" {
		log.Fatal("MONGO_DB_NAME environment variable is not set")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	MongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB client:", err)
	}

	// Ping database untuk verifikasi
	if err = MongoClient.Ping(ctx, nil); err != nil {
		log.Fatal("Error pinging MongoDB:", err)
	}

	MongoDB = MongoClient.Database(dbName)

	log.Println("Connected to MongoDB database:", dbName)
}

func CloseDB() {
	if PgDB != nil {
		if err := PgDB.Close(); err != nil {
			log.Println("Error closing Postgres:", err)
		} else {
			log.Println("Postgres connection closed")
		}
	}

	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := MongoClient.Disconnect(ctx); err != nil {
			log.Println("Error disconnecting MongoDB:", err)
		} else {
			log.Println("MongoDB connection closed")
		}
	}
}