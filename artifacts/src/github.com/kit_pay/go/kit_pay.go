package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

)

// SmartContract Define the Smart Contract structure
// SmartContract 스마트계약의 구조를 정의
type SmartContract struct {
}

//Point 포인트의 구조를 정의 구조 태그는 json라이브러리의 인코딩에 사용된다.
type Point struct {
	Owner string `json:"owner"`
}

//거래 정보의 구조를 정의 
type History struct {
	Sender string `json:"sender"`
	Receiver string `json:"receiver"`
	Amount string `json:"amount"`
	Date string `json:"date"`
}

type Result struct {
	Amount string `json: amount`
}


//스마트 컨트랙트를 초기화하는 메소드
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("kit_pay_cc")

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "createPoint":
		return s.createPoint(APIstub, args)
	case "deletePoint":
		return s.deletePoint(APIstub, args)
	case "queryPoint":
		return s.queryPoint(APIstub, args)
	case "changePointOwner":
		return s.changePointOwner(APIstub, args)
	case "getHistory":
		return s.getHistory(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}

	// return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) createPoint(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	var i int

	var point = Point{}
	indexName := "point~key"

	point = Point{Owner: args[0]}
	cnt, _ := strconv.Atoi(args[1])

	for i = 0; i < cnt; i++ {
		date := time.Now().Format("2006-01-02 15:04:05")
		key := "point " + date + strconv.Itoa(i)
		pointAsBytes,_ := json.Marshal(point)
	
		APIstub.PutState(key, pointAsBytes)
		colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{point.Owner, key})
		if err != nil {
			return shim.Error(err.Error())
		}

		value := []byte{0x00}
		APIstub.PutState(colorNameIndexKey, value)	
	}

	return shim.Success(nil)
}

func (s *SmartContract) deletePoint(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//상품권 삭제
	//APIstub.getStateByPartialCompositeKey로 찾을 리스트 를 뽑은 다음 전달받은 금액만큼을 삭제(컴포지트 키 사용)
	//iterator를 반환, 그런 다음 for문을 실행, DelState로 기존 토큰뿐만 아니라 복합키 자료도 폐기해야 한다.

	if len(args) != 2 {//args = {managerID, amount}
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	owner := args[0]
	cnt, _ := strconv.Atoi(args[1])

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("owner~key", []string{owner})
	//StateQueryIteratorInterface를 반환, key/value의 형태로 사용
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	for i = 0; i < cnt; i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)//복호화 키, 복호화 후 배열, 에러
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[0]

		APIstub.DelState(id)//Point 원장 삭제
		APIstub.DelState(responseRange.Key)//복합키 원장 삭제

		fmt.Printf("Delete a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
	}

	return shim.Success(nil)
}

func (s *SmartContract) queryPoint(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//일단 밑의 queryPoint로 상품권 조회 호출

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments 1")
	}
	owner := args[0]

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("point~key", []string{owner})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var cnt int
	var id string

	var points []byte
	bArrayMemberAlreadyWritten := false

	cnt = 0

	points = append([]byte("["))

	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]//key
		assetAsBytes, err := APIstub.GetState(id)

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			points = append(points, newBytes...)

		} else {
			// newBytes := append([]byte(","), pointsAsBytes...)
			points = append(points, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

		cnt++

	}

	var result = Result{}
	result = Result{Amount: strconv.Itoa(cnt)}

	resultAsBytes, _ := json.Marshal(result)
	newBytes := append([]byte(","), resultAsBytes...)
	points = append(points, newBytes...)
	points = append(points, []byte("]")...)

	return shim.Success(points)
}

func (s *SmartContract) changePointOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {//sender, receiver, amount
		return shim.Error("Incorrect number of arguments")
	}
	sender := args[0]
	receiver := args[1]
	cnt,_ := strconv.Atoi(args[2])

	indexName := "point~key"

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("point~key", []string{sender})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	point := Point{}
	history := History{}

	//포인트 생성
	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		if i == cnt {
			break;
		}
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		//토큰 변경 시작
		id = compositeKeyParts[1]//key
		pointAsBytes, err := APIstub.GetState(id)
		
		json.Unmarshal(pointAsBytes, &point)
		point.Owner = receiver//주인 변경

		pointAsBytes, _ = json.Marshal(point)
		APIstub.PutState(id, pointAsBytes);//변경 토큰 저장

		//delete old composite key
		APIstub.DelState(responseRange.Key)

		//create new composite key
		colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{receiver, id})
		value := []byte{0x00}
		APIstub.PutState(colorNameIndexKey, value)

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
	}

	var key string

	//거래내역 생성
	date := time.Now().Format("2006-01-02 15:04:05")
	history = History{Sender: sender, Receiver: receiver, Amount: args[2], Date: date}

	historyAsBytes,_  := json.Marshal(history)
	key = sender + receiver + date
	APIstub.PutState(key, historyAsBytes)

	indexName = "history~key"

	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{sender, receiver, key})
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	return shim.Success(historyAsBytes)
}

func (s *SmartContract) getHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {//찾으려는 사람
		return shim.Error("Incorrect number of arguments")
	}
	owner := args[0]

	indexName := "history~key"

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey(indexName, []string{owner})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	var historys []byte
	bArrayMemberAlreadyWritten := false

	historys = append([]byte("["))

	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[2]//sender, receiver, key
		assetAsBytes, err := APIstub.GetState(id)

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			historys = append(historys, newBytes...)

		} else {
			historys = append(historys, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	historys = append(historys, []byte("]")...)

	return shim.Success(historys)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	points := []Point{
		Point{Owner: "park"},
		Point{Owner: "park"},
		Point{Owner: "park"},
		Point{Owner: "park"},
		Point{Owner: "park"},
		Point{Owner: "kim"},
		Point{Owner: "kim"},
		Point{Owner: "kim"},
		Point{Owner: "kim"},
		Point{Owner: "kim"},
		Point{Owner: "son"},
		Point{Owner: "son"},
		Point{Owner: "son"},
		Point{Owner: "son"},
		Point{Owner: "son"},
	}

	i := 0
	for i < len(points) {
		pointAsBytes, _ := json.Marshal(points[i])
		Key := "POINT"+strconv.Itoa(i)
		APIstub.PutState(Key, pointAsBytes)

		indexName := "owner~key"//복합키 작성을 위한 문자열.
		colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{points[i].Owner, Key})//복합키 작성
		if err != nil {
			return shim.Error(err.Error())
		}

		value := []byte{0x00}//값을 저장하기 위해 만든 빈 값
		APIstub.PutState(colorNameIndexKey, value)//만들어 놓은 복합 키, 빈 값 복합키를 저장하는 것 자체가 목적

		i = i + 1
	}

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
