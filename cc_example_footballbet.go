package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"crypto/sha1"
	"encoding/hex"
)

type SimpleChaincode struct {
}

type FootballBet struct {
	IdClient    string `json:"idClient"`
	Timestamp    string `json:"timestamp"`
	HomeTeamId     string `json:"homeTeamId"`
	AwayTeamId    string `json:"awayTeamId"`
	HomeTeamScore    string `json:"homeTeamScore"`
	AwayTeamScore        string `json:"awayTeamScore"`
	Currency     string `json:"currency"`
	HomeTeamOdds string    `json:"homeTeamOdds"`
	AwayTeamOdds string `json:"awayTeamOdds"`
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Printf("\nInit OK\n")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Invoke
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemented in invoke
		return t.query(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

//Handling unknown invoke
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//opt1: inserTrackInfo
	if args[0] == "insertFootballBet" {
		return t.insertFootballBet(stub, args)
	} else {
		fmt.Printf("\nCommand invoke not found, different from insertFootballBet\n")
		return shim.Success(nil)
	}
}

// Deletes an entity from state. Not implemented for this version
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//Not implemented for this version
	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if args[0] == "queryFootballBet" {
		return t.queryBetsByIdClient(stub, args)
	} else {
		fmt.Printf("\nCommand query  not found, different from queryFootballBet\n")
		return shim.Success(nil)
	}

}

func (t *SimpleChaincode) insertFootballBet(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}

	var err error

	_idClient := strings.ToLower(args[1])
	_timestamp := strings.ToLower(args[2])
	_homeTeamId := strings.ToLower(args[3])
	_awayTeamId := strings.ToLower(args[4])
	_homeTeamScore := strings.ToLower(args[5])
	_awayTeamScore := strings.ToLower(args[6])
	_currency := strings.ToLower(args[7])
	_homeTeamOdds := strings.ToLower(args[8])
	_awayTeamOdds := strings.ToLower(args[9])

	footballBet := &FootballBet{_idClient,
					 _timestamp, 
					 _homeTeamId,
					 _awayTeamId, 
					 _homeTeamScore,
					 _awayTeamScore,
					 _currency,
					 _homeTeamOdds,
					 _awayTeamOdds}

	footballBetAsBytes, errorMarshalling := json.Marshal(footballBet)

	if errorMarshalling != nil {
			return shim.Error(errorMarshalling.Error())
		}

	// === Save info to state ===
	err = stub.PutState(args[0]+"_"+_idClient+"_"+getHash(_timestamp), footballBetAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Creating index ===
	indexName := "idClient~timestamp"
	footballBetNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{footballBet.IdClient, footballBet.Timestamp})
	if err != nil {
		return shim.Error(err.Error())
	}



	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(footballBetNameIndexKey, value)

	// ==== Track saved and indexed. Return success ====
	return shim.Success(nil)

}

func getHash(text string) string{
	hash := sha1.New()
	hash.Write([]byte(text))
	sha1_hash := hex.EncodeToString(hash.Sum(nil))
	return sha1_hash
}

func (t *SimpleChaincode) queryBetsByIdClient(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	idClient := args[1]

	queryString := "{\"selector\":{\"idClient\":{ \"$eq\":\"" + idClient + "\" } " + "}}"

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	resultsIterator, err1 := stub.GetQueryResult(queryString)
	if err1 != nil {
		return nil, err1
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	for resultsIterator.HasNext() {
		queryResponse, err2 := resultsIterator.Next()
		if err2 != nil {
			return nil, err2
		}
		buffer.WriteString(string(queryResponse))
	}
	return buffer.Bytes(), nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
