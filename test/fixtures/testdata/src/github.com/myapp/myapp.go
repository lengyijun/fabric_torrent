package main

import (
    "bytes"
    "encoding/json"
    "encoding/pem"
    "fmt"
    "crypto/x509"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {

}

type File struct {
    Name string `json:"name"`
    Hash string `json:"hash"`
    Keyword string `json:"keyword"`
    Summary string `json:"summary"`
    Owner string `json:"owner"`
    Locktime int64 `json:"locktime"`
    Magnet string
    AESKey string `tempory`
}

/*
* Init function: necessary
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
    return shim.Success(nil)
}


/*
 * Invoke function: necessary
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

    function, args := APIstub.GetFunctionAndParameters()
    if function == "createFile" {
        return s.createFile(APIstub, args)
    } else if function == "queryFile" {
        return s.queryFile(APIstub, args)
    } else if function == "changeFileOwner" {
        return s.changeFileOwner(APIstub, args)
    } else if function == "deleteFile" {
        return s.deleteFile(APIstub, args)
    } else if function == "externalTestLocktime" {
        return s.externalTestLocktime(APIstub, args)
    } else if function == "addLocktime" {
        return s.addLocktime(APIstub, args)
    } else if function == "getAllMagnet"{
        return s.getAllMagnet(APIstub)
    }

    return shim.Error("Invalid Smart Contract function name.")
}


/*
 * createFile function: create a new file record by providing key values
 */
func (s *SmartContract) createFile(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    uname, err := s.testCertificate(APIstub, nil)
    if err != nil {
        return shim.Error(err.Error())
    }

    //check if exist a file with same name
    //haven't tested
    resultsIterator, err := APIstub.GetStateByPartialCompositeKey("File",[]string{args[0]})
    if err != nil {
        return shim.Error(err.Error())
    }
    defer resultsIterator.Close()

    if resultsIterator.HasNext(){
        return shim.Error("already exist a file having the same name")
    }

    // create an object
    //var file = File{Name: args[0], Hash: args[1], Keyword: args[2], Summary: args[3], Owner: uname, Locktime: 0,Magnet:args[4]}
    var file = File{Name: args[0], Hash: args[1], Keyword: args[2], Summary: args[3], Owner: uname, Locktime: 0,Magnet:args[4],AESKey:args[5]}
    fileAsBytes, _ := json.Marshal(file)

    // we need a relational database as an addition to leveldb
    // edit here when custom interface is ready
    // we currently use composite key with keyword name and owner.

    //args[2]: keyword
    //args[0]: Name
    //uname:
    keys := []string{args[2], args[0], uname}
    ckey, err := APIstub.CreateCompositeKey("File", keys)
    if err != nil {
        return shim.Error(err.Error())
    }
    APIstub.PutState(ckey, fileAsBytes)

    APIstub.SetEvent("createFile", fileAsBytes);
    return shim.Success([]byte(uname))
}


/*
 *queryFile function: query File by at least one at most three keys
 */
func (s *SmartContract) queryFile(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    if len(args) > 3 {
        return shim.Error("Incorrect number of arguments. Expacting at least one key from name, keyword and owner to search file system")
    }

    // get query result
    resultsIterator, err := APIstub.GetStateByPartialCompositeKey("File", args)
    if err != nil {
        return shim.Error(err.Error())
    }

    defer resultsIterator.Close()

    // buffer is a JSON array containing query results
    //var buffer bytes.Buffer
    //buffer.WriteString("[")

    //bArrayMemberAlreadyWritten := false

    if resultsIterator.HasNext() {
    	kv,_:=resultsIterator.Next()
    	file:=File{}
    	json.Unmarshal(kv.Value,&file)
    	res:=make([][]byte,2)
        res[0]=[]byte(file.Name)
    	res[1]=[]byte(file.AESKey)
    	return shim.Success(bytes.Join(res,[]byte(",")))
        //queryResponse, err := resultsIterator.Next()
        //if err != nil {
        //    return shim.Error(err.Error())
        //}
        //if bArrayMemberAlreadyWritten == true {
        //    buffer.WriteString(",")
        //}
        //buffer.WriteString("{\"Key\":{\"objectType\":")
        //typeString, keys, err := APIstub.SplitCompositeKey(queryResponse.Key)
        //if err != nil {
        //    return shim.Error(err.Error())
        //}
        //buffer.WriteString(typeString)
        //buffer.WriteString("\", \"attributes\":[\"")
        //buffer.WriteString(strings.Join(keys, "\", \""))
        //buffer.WriteString("\"]}, \"Record\":")
        //buffer.WriteString(string(queryResponse.Value))
        //buffer.WriteString("}")
        //bArrayMemberAlreadyWritten = true
    }
    //buffer.WriteString("]")

    return shim.Error("no key")
    //return shim.Success(buffer.Bytes())
}


/*
 * changeFileOwner function: change owner of a file. must provide complete composite key
 */
func (s *SmartContract) changeFileOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 4 {
        return shim.Error("Incorrect number of arguments. Expecting 3 keys and 1 new owner")
    }

    uname, err := s.testCertificate(APIstub, nil)
    if err != nil {
        return shim.Error(err.Error())
    }

    // test Locktime
    timeflag := s.testLocktime(APIstub, []string{args[0], args[1], args[2]})
    if timeflag >= 2 {
        return shim.Error("The file is locked")
    }

    // create composite key
    keys := []string{args[0], args[1], args[2]}
    ckey, err := APIstub.CreateCompositeKey("File", keys)
    if err != nil {
        return shim.Error(err.Error())
    }
    //query the File
    fileAsBytes, _ := APIstub.GetState(ckey)
    file := File{}
    json.Unmarshal(fileAsBytes, &file)

    if file.Owner != uname {
        return shim.Error("Permission denied")
    }

    // edit Owner attribute
    file.Owner = args[3]
    fileAsBytes, _ = json.Marshal(file)
    APIstub.PutState(ckey, fileAsBytes)

    return shim.Success(nil)
}


/*
 * deleteFile function: delete the whole file.  must provide complete composite key
 */
func (s *SmartContract) deleteFile(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 3 {
        return shim.Error("Incorrect number of arguments. Expecting 3 keys")
    }

    // test Locktime
    timeflag := s.testLocktime(APIstub, []string{args[0], args[1], args[2]})
    if timeflag >= 2 {
        return shim.Error("The file is locked")
    }

    uname, err := s.testCertificate(APIstub, nil)
    if err != nil {
        return shim.Error(err.Error())
    }
    if uname != args[2] {
        return shim.Error("permission denied")
    }

    // create composite key
    keys := []string{args[0], args[1], args[2]}
    ckey, err := APIstub.CreateCompositeKey("File", keys)
    if err != nil {
        return shim.Error(err.Error())
    }
    //query the File
    err = APIstub.DelState(ckey)
    if err != nil {
        return shim.Error(err.Error())
    }

    APIstub.SetEvent("deleteFile", []byte(ckey));
    return shim.Success([]byte(uname))
}


/*
 * addLocktime function: called by exchange chaincode when the file owner respond file request
 */
func (s *SmartContract) addLocktime(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting file key")
    }

    // check certificate
    uname, err := s.testCertificate(APIstub, nil)
    if err != nil {
        return shim.Error(err.Error())
    }

    // test Locktime
    timeflag := s.testLocktime(APIstub, []string{args[0]})
    if timeflag != 0 {
        return shim.Error("The file is locked")
    }

    // get Tx Timestamp
    timestamp, err := APIstub.GetTxTimestamp()
    if err != nil {
        return shim.Error(err.Error())
    }

    //query the File
    fileAsBytes, _ := APIstub.GetState(args[0])
    file := File{}
    json.Unmarshal(fileAsBytes, &file)

    if file.Owner != uname {
        return shim.Error("Permission denied")
    }
    // edit Locktime attribute
    file.Locktime = timestamp.GetSeconds()
    fileAsBytes, _ = json.Marshal(file)
    APIstub.PutState(args[0], fileAsBytes)

    return shim.Success(nil)

}


func (s *SmartContract) testLocktime(APIstub shim.ChaincodeStubInterface, args []string) (int) {

    if len(args) != 1 {
        return 3
    }

    //query the File
    fileAsBytes, _ := APIstub.GetState(args[0])
    file := File{}
    json.Unmarshal(fileAsBytes, &file)

    timestamp, err := APIstub.GetTxTimestamp()
    if err != nil {
        return 4
    }

    intTimestamp := timestamp.GetSeconds()
    fileTimestamp := file.Locktime

    // 2 - waiting for file requester to confirm
    // 1 - Comfirm time out. Free for file owner to edit the file
    // 0 - normal time
    // both 1 & 2 the file cannot be requested again
    if intTimestamp <= fileTimestamp + 300 {
        return 2
    } else if intTimestamp <= fileTimestamp + 600 {
        return 1
    } else {
        return 0
    }
}


/*
 * externalTestLocktime function: Different from testLocktime, this function is called by exchange
 * chaincode
 */
func (s *SmartContract) externalTestLocktime(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting file key")
    }

    //query the File
    fileAsBytes, _ := APIstub.GetState(args[0])
    file := File{}
    json.Unmarshal(fileAsBytes, &file)

    timestamp, err := APIstub.GetTxTimestamp()
    if err != nil {
        return shim.Error(err.Error())
    }

    intTimestamp := timestamp.GetSeconds()
    fileTimestamp := file.Locktime

    // 2 - waiting for file requester to confirm
    // 1 - Comfirm time out. Free for file owner to edit the file
    // 0 - normal time
    // both 1 & 2 the file cannot be requested again
    if intTimestamp <= fileTimestamp + 300 {
        return shim.Success([]byte("2"))
    } else if intTimestamp <= fileTimestamp + 600 {
        return shim.Success([]byte("1"))
    } else {
        return shim.Success([]byte("0"))
    }
}


func (s *SmartContract) testCertificate(stub shim.ChaincodeStubInterface, args []string ) (string, error) {
    creatorByte, _ := stub.GetCreator()
    certStart := bytes.IndexAny(creatorByte, "-----BEGIN")
    if certStart == -1 {
        return "", fmt.Errorf("%s", "no certificate detected")
    }

    certText := creatorByte[certStart:]
    content, _ := pem.Decode(certText)
    if content == nil {
        return "", fmt.Errorf("%s", "fail to decode the certificate")
    }

    cert, err := x509.ParseCertificate(content.Bytes)
    if err != nil {
        return "", fmt.Errorf("%s", "fail when parsing the x509 certificate")
    }
    cname := cert.Subject.CommonName
    return cname, nil
}

func (s *SmartContract) getAllMagnet(stub shim.ChaincodeStubInterface) sc.Response{
    keysIter, err := stub.GetStateByRange("","")
    if err != nil {
        return shim.Error(fmt.Sprintf("keys operation failed. Error accessing state: %s", err))
    }
    defer keysIter.Close()

    var m=make([][]byte,0)
    for keysIter.HasNext() {
        file := File{}
        value, iterErr := keysIter.Next()
        if iterErr != nil {
            return shim.Error(fmt.Sprintf("keys operation failed. Error accessing state: %s", err))
        }
        json.Unmarshal(value.Value,&file)
        fmt.Println(file.Magnet)
        m=append(m, []byte(file.Magnet))
    }
    return shim.Success(bytes.Join(m,[]byte(",,")))

}

// for test
func main() {
    err := shim.Start(new(SmartContract))
    if err != nil {
        fmt.Printf("Error creating new Smart Contract: %s", err)
    }
}
