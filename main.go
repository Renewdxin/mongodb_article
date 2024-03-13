package main

import (
	"context"
	_ "fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

// Article represents the schema for the "articles" collection
type Article struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Title   string             `bson:"title"`
	Content string             `bson:"content"`
	Author  string             `bson:"author"`
}

var (
	ctx        context.Context
	collection *mongo.Collection
)

func init() {
	// Initialize MongoDB client
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Ping the primary
	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}

	// Get a handle for your collection
	collection = client.Database("article").Collection("article")
}

func main() {
	router := gin.Default()

	router.POST("/articles", createArticle)
	router.GET("/articles/:id", getArticle)
	router.GET("/articles", listArticles)
	router.PUT("/articles/:id", updateArticle)
	router.DELETE("/articles/:id", deleteArticle)

	router.Run(":8080")
}

func createArticle(c *gin.Context) {
	var article Article
	if err := c.BindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := collection.InsertOne(ctx, article)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.InsertedID})
}

func getArticle(c *gin.Context) {
	id := c.Param("id")

	objID, _ := primitive.ObjectIDFromHex(id)
	var article Article
	err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&article)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, article)
}

func listArticles(c *gin.Context) {
	// 设置投影，排除_id字段
	projection := bson.D{{"_id", 0}}
	opts := options.Find().SetProjection(projection)

	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var articles []bson.M
	if err = cursor.All(context.Background(), &articles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回结果时不包含_id字段
	c.JSON(http.StatusOK, articles)
}

func updateArticle(c *gin.Context) {
	id := c.Param("id")

	objID, _ := primitive.ObjectIDFromHex(id)
	var article Article
	if err := c.BindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": article})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": result.ModifiedCount})
}

func deleteArticle(c *gin.Context) {
	id := c.Param("id")

	objID, _ := primitive.ObjectIDFromHex(id)
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": result.DeletedCount})
}
