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

type Request struct {
    From string `json:"from"`
    To string `json:"to"`
    File string `json:"file"`
    RequestTime int64 `json:"requestTime"`
    ResponseTime int64 `json:"responseTime"`
    ConfirmationTime int64 `json:"confirmationTime"`
}

type RequestMessage struct {
    From string `json:"from"`
    To string `json:"to"`
    File string `json:"file"`
    TxID string `json:"tx_id"`
    RequestTime int64 `json:"requestTime"`
}

type ResponseMessage struct {
    From string `json:"from"`
    To []string `json:"to"`
    File string `json:"file"`
    TxID []string `json:"tx_id"`
    Secret string `json:"secret"`
    ResponseTime int64 `json:responseTime`
}

type ConfirmationMessage struct {
    TxID string `json:"tx_id"`
    ConfirmationTime int64 `json:"confirmationTime"`
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
    if function == "requestSecret" {
        return s.requestSecret(APIstub, args)
    } else if function == "respondSecret" {
        return s.respondSecret(APIstub, args)
    } else if function == "confirmSecret" {
        return s.confirmSecret(APIstub, args)
    } else if function == "queryRequest" {
        return s.queryRequest(APIstub, args)
    }

    return shim.Error("Invalid Smart Contract function name.")
}


func (s *SmartContract) requestSecret(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    if len(args) != 3 {
        return shim.Error("Incorrect number of arguments. Expecting 3 keys of file")
    }

    uname, err := s.testCertificate(APIstub, nil)
    if err != nil {
        return shim.Error(err.Error())
    }
    // produce the composite key for file
    keys := []string{args[0], args[1], args[2]}
    ckey, err := APIstub.CreateCompositeKey("File", keys)
    if err != nil {
        return shim.Error(err.Error())
    }

    argsByBytes := [][]byte{[]byte("externalTestLocktime"), []byte(ckey)}
    res := APIstub.InvokeChaincode("myapp", argsByBytes, "")
    if res.Status > 400 {
        return shim.Error(res.Message)
    }
    if len(res.Payload) <= 0 {
        return shim.Error("The file is not exist")
    } else {
        ret := string(res.Payload)
        if ret != "0" {
            return shim.Error("The file is locked for request")
        }
    }

    // check the existence of the file
    argsByBytes = [][]byte{[]byte("queryFile"), []byte(args[0]), []byte(args[1]), []byte(args[2])}
    res = APIstub.InvokeChaincode("myapp", argsByBytes, "")
    if res.Status > 400 {
        return shim.Error("Fail to call file chaincode")
    }
    //if len(res.Payload) <= 2 {
    //    return shim.Error("The file is not exist")
    //}

    // get timestamp and tx_id
    tx_id := APIstub.GetTxID()
    timestamp, err := APIstub.GetTxTimestamp()
    if err != nil {
        return shim.Error(err.Error())
    }

    // put request record
    var request = Request{From: uname, To: args[2], File: ckey, RequestTime: timestamp.GetSeconds(), ResponseTime: 0, ConfirmationTime: 0}
    requestAsBytes, _ := json.Marshal(request)

    APIstub.PutState(tx_id, requestAsBytes)

    // broadcast an event
    var message = RequestMessage{From: uname, To: args[2], File: ckey, TxID: tx_id, RequestTime: timestamp.GetSeconds()}
    messageAsBytes, _ := json.Marshal(message)
    APIstub.SetEvent("requestSecret", messageAsBytes)

    return shim.Success(res.Payload)//todo
}


func (s *SmartContract) respondSecret(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    if len(args) <= 1 {
        return shim.Error("Incorrect number of arguments. Expecting at least one tx_id and one secret")
    }

    uname, err := s.testCertificate(APIstub, nil)
    if err != nil {
        return shim.Error(err.Error())
    }

    var fileKey = ""
    var fromList []string
    var timestampInt int64

    for _, req := range args[:len(args)-1] {
        // get the request record by tx_id
        requestAsBytes, err := APIstub.GetState(req)
        request := Request{}
        json.Unmarshal(requestAsBytes, &request)

        if fileKey == "" {
            fileKey = request.File
        } else if fileKey != request.File {
            return shim.Error("Wrong request tx_id: Inconsist file")
        }

        // check
        if uname != request.To {
            return shim.Error("Wrong transaction ID")
        }

        fromList = append(fromList, request.From)

        // add timestamp
        timestamp, err := APIstub.GetTxTimestamp()
        if err != nil {
            return shim.Error(err.Error())
        }
        if request.ResponseTime == 0 {
            timestampInt = timestamp.GetSeconds()
            request.ResponseTime = timestampInt
        } else {
            return shim.Error("This request already has a response")
        }
        requestAsBytes, _ = json.Marshal(request)
        APIstub.PutState(args[0], requestAsBytes)
    }

    argsByBytes := [][]byte{[]byte("addLocktime"), []byte(fileKey)}
    res := APIstub.InvokeChaincode("myapp", argsByBytes, "")
    if res.Status > 400 {
        return shim.Error(res.Message)
    }

    // broadcast an event
    var message = ResponseMessage{From: uname, To: fromList, File: fileKey, TxID: args[:len(args)-1], Secret: args[len(args)-1], ResponseTime: timestampInt}
    messageAsBytes, _ := json.Marshal(message)
    APIstub.SetEvent("respondSecret", messageAsBytes)

    return shim.Success(nil)
}


func (s *SmartContract) confirmSecret(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting only tx_id")
    }

    uname, err := s.testCertificate(APIstub, nil)
    if err != nil {
        return shim.Error(err.Error())
    }

    // get the request record by tx_id
    requestAsBytes, err := APIstub.GetState(args[0])
    request := Request{}
    json.Unmarshal(requestAsBytes, &request)

    // check
    if uname != request.From {
        return shim.Error("Wrong transaction ID")
    }

    // test Locktime
    argsByBytes := [][]byte{[]byte("externalTestLocktime"), []byte(request.File)}
    res := APIstub.InvokeChaincode("myapp", argsByBytes, "")
    if res.Status > 400 {
        return shim.Error(res.Message)
    } else if len(res.Payload) <= 0 {
        return shim.Error("The file is not exist")
    } else {
        ret := string(res.Payload)
        if ret != "2" {
            return shim.Error("The file is locked for confirm")
        }
    }

    // add timestamp
    timestamp, err := APIstub.GetTxTimestamp()
    if err != nil {
        return shim.Error(err.Error())
    }
    if request.ConfirmationTime == 0 {
        request.ConfirmationTime = timestamp.GetSeconds()
    } else {
        return shim.Error("This request has been confirmed")
    }
    requestAsBytes, _ = json.Marshal(request)
    APIstub.PutState(args[0], requestAsBytes)

    var message = ConfirmationMessage{TxID: args[0], ConfirmationTime: timestamp.GetSeconds()} 
    messageAsBytes, _ := json.Marshal(message)
    APIstub.SetEvent("confirmSecret", messageAsBytes)

    return shim.Success(nil)
}


func (s *SmartContract) queryRequest(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    if len(args) > 1 {
        return shim.Error("Incorrect number of arguments. Expacting only transaction id")
    }

    // buffer is a JSON array containing query results
    var buffer bytes.Buffer

    queryResponse, err := APIstub.GetState(args[0])
    if err != nil {
        return shim.Error(err.Error())
    }
    buffer.WriteString("{\"Key\":\"")
    if err != nil {
        return shim.Error(err.Error())
    }
    buffer.WriteString(args[0])
    buffer.WriteString("\", \"Record\":")
    buffer.WriteString(string(queryResponse))
    buffer.WriteString("}")

    return shim.Success(buffer.Bytes())
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


// for test
func main() {
    err := shim.Start(new(SmartContract))
    if err != nil {
        fmt.Printf("Error creating new Smart Contract: %s", err)
    }
}
