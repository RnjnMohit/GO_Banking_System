package controllers

import (
	"banking/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const connectionString = "mongodb+srv://votingapp:Ranjan11082@clustervote.ccp4sos.mongodb.net/banking"
const dbName = "banking"
const colName = "account"

var collection *mongo.Collection

func init() {
	clientOption := options.Client().ApplyURI(connectionString)
	
	client, err := mongo.Connect(context.TODO(), clientOption)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("MongoDB Connection Successful")
	collection = client.Database(dbName).Collection(colName)
	fmt.Println("Collection is ready")
}

func createAccount(account models.Account) error {
	existingAccount := models.Account{}
	err := collection.FindOne(context.Background(), bson.M{"security.nickname": account.Security.NickName}).Decode(&existingAccount)

	if err == nil {
		return fmt.Errorf("account with nickname %s already exists", account.Security.NickName)
	} else if err != mongo.ErrNoDocuments {
		log.Printf("Error checking for existing account: %v", err)
		return err
	}

	if account.ID.IsZero() {
		account.ID = primitive.NewObjectID()
	}

	//hashing the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Security.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return err
	}
	account.Security.Password = string(hashedPassword)

	insertResult, err := collection.InsertOne(context.Background(), account)
	if err != nil {
		log.Printf("Could not create account: %v", err)
		return err
	}

	fmt.Printf("Inserted a single document: %v\n", insertResult.InsertedID)
	return nil
}

func getAllAccount() ([]models.Account, error){
	filter := bson.D{{}}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil{
		log.Printf("Error fetching Account: %v", err)
		return nil, err
	}

	defer cursor.Close(context.Background())

	var accounts []models.Account
	for cursor.Next(context.Background()){
		var account models.Account
		if err:= cursor.Decode(&account); err!=nil{
			log.Printf("Error decoding account document: %v", err)
			return nil, err
		}

		accounts = append(accounts, account)
	}

	if err := cursor.Err(); err != nil{
		log.Printf("Error during curson iteration : %v", err)
		return nil, err
	}
	return accounts, nil
}

func getOneAccount(accountID primitive.ObjectID) (*models.Account, error){
	var account models.Account
	err := collection.FindOne(context.Background(), bson.M{"_id": accountID}).Decode(&account)
	if err!=nil{
		if err == mongo.ErrNoDocuments{
			return nil,fmt.Errorf("No account found with ID %s", accountID.Hex())
		}
		return nil, err
	}
	return &account, nil
  
}

func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var account models.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		http.Error(w, "Cannot parse JSON", http.StatusBadRequest)
		return
	}

	if err := createAccount(account); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(account); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func GetAllAccountHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "Invalid request Method", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	accounts, err := getAllAccount()
	if err != nil{
		http.Error(w, "Failed to fetch account", http.StatusInternalServerError)
		return
	}
	for _, acc := range accounts{
		fmt.Println("Accound ID: ", acc.ID.Hex())
		fmt.Println("Account Mmeber Name: ", acc.Security.NickName)
	}

	if err := json.NewEncoder(w).Encode(accounts); err != nil{
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func GetOneAccountHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "Invalid request Method", http.StatusMethodNotAllowed)
		return
	}

	//THIS IS USING QUERY PARAMETE---->>>
	// ac:countIDHex := r.URL.Query().Get("id")

	//HERE USING PATH PARAMETER-->
	vars := mux.Vars(r)
	accountIDHex := vars["id"]
	if accountIDHex == ""{
		http.Error(w, "Missing account ID", http.StatusBadRequest)
		return
	}

	accountID, err := primitive.ObjectIDFromHex(accountIDHex)
	if err != nil{
		http.Error(w, "Invalid account ID format", http.StatusBadRequest)
		return
	}

	account, err := getOneAccount(accountID)
	if err!= nil{
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(account); err != nil{
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodDelete{
		http.Error(w, "Invalid request Method", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	accountIDHex := vars["id"]

	if accountIDHex == ""{
		http.Error(w, "Missing account ID", http.StatusBadRequest)
		return 
	}

	accountID, err := primitive.ObjectIDFromHex(accountIDHex)
	if err != nil{
		http.Error(w, "Invalid account id format", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": accountID}
	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil{
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}
	fmt.Println("Account Successfully deleted")

	if deleteResult.DeletedCount == 0{
		http.Error(w, "No account found with the give ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteAllAccountsHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    filter := bson.D{} // Empty filter matches all documents
    deleteResult, err := collection.DeleteMany(context.Background(), filter)
    if err != nil {
        http.Error(w, "Failed to delete accounts", http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "deleted_count": deleteResult.DeletedCount,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to write response", http.StatusInternalServerError)
    }
}