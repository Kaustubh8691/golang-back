package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	helper "github.com/Kaustubh8691/golang-backend/helpers"
	"github.com/Kaustubh8691/golang-backend/models"

	"github.com/Kaustubh8691/golang-backend/database"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var dataCollection *mongo.Collection = database.OpenCollection(database.Client, "data")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}

	return check, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:8080")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for phone"})
		}
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already register"})
		}
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		data := user.ID.Hex()
		user.User_id = data
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.User_type, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})

		}
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.User_type, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUSerType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}
		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$Root"}}}}}}

		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}}}}
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listting user items"})

		}
		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allUsers[0])
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		cursor, err := userCollection.Find(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}
		var user []bson.M
		if err = cursor.All(ctx, &user); err != nil {
			log.Fatal(err)
		}
		fmt.Println(user)

		defer cancel()

		c.JSON(http.StatusOK, user)

	}
}
func Crea() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var data models.Data

		if err := c.BindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error1": err.Error()})
			return
		}
		validationErr := validate.Struct(data)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errorss": validationErr.Error()})
			return
		}
		count, err := dataCollection.CountDocuments(ctx, bson.M{"name": data.Name})
		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro2r": "error occured while checking for the email"})
		}
		data.Count = string(count)
		data.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		data.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		data.ID = primitive.NewObjectID()
		data1 := data.ID.Hex()
		data.User_id = data1

		resultInsertionNumber, insertErr := dataCollection.InsertOne(ctx, data)
		if insertErr != nil {
			msg := fmt.Sprintf("data not created")
			c.JSON(http.StatusInternalServerError, gin.H{"errorg": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

func GetData() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		cursor, err := dataCollection.Find(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}
		var data []bson.M
		if err = cursor.All(ctx, &data); err != nil {
			log.Fatal(err)
		}
		fmt.Println(data)

		defer cancel()

		c.JSON(http.StatusOK, data)

	}
}

func UpdateData() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var data models.Data
		err := dataCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&data)
		// err := dataCollection.FindOneAndUpdate(ctx, bson.M{"user_id": userId},"$set"{"user_id": userId}).Decode(&data)
		if errr := c.ShouldBindUri(&data); errr == nil {
			fmt.Printf("data - %+v", data)
		} else {
			fmt.Printf("error - %+v", errr)
		}
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)

	}
}

func DeleteData() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		// if err := helper.MatchUserTypeToid(c, userId); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return
		// }
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var data models.Data
		err := dataCollection.FindOneAndDelete(ctx, bson.M{"user_id": userId}).Decode(&data)
		// err := dataCollection.FindOneAndUpdate(ctx, bson.M{"user_id": userId},"$set"{"user_id": userId}).Decode(&data)

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)

		// 	fmt.Println("3. Performing Http Put...")
		// todo := Todo{1, 2, "lorem ipsum dolor sit amet", true}
		// jsonReq, err := json.Marshal(todo)
		// req, err := http.NewRequest(http.MethodPut, "https://jsonplaceholder.typicode.com/todos/1", bytes.NewBuffer(jsonReq))
		// req.Header.Set("Content-Type", "application/json; charset=utf-8")
		// client := &http.Client{}
		// resp, err := client.Do(req)
		// if err != nil {
		//     log.Fatalln(err)
		// }

		// defer resp.Body.Close()
		// bodyBytes, _ := ioutil.ReadAll(resp.Body)

		// // Convert response body to string
		// bodyString := string(bodyBytes)
		// fmt.Println(bodyString)

		// // Convert response body to Todo struct
		// var todoStruct Todo
		// json.Unmarshal(bodyBytes, &todoStruct)
		// fmt.Printf("API Response as struct:\n%+v\n", todoStruct)
	}
}
