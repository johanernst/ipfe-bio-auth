package main

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestRuntime(t *testing.T) {
	fmt.Println(os.Args)

	numberOfClients, err := strconv.Atoi(os.Args[3])
	exitOnError(err)
	templateLength, err := strconv.Atoi(os.Args[4])
	exitOnError(err)
	maxValueTemplate, err := strconv.Atoi(os.Args[5])
	exitOnError(err)
	ipfeTiming = (os.Args[6] == "ipfeTiming")
	var maxValueIpfe = int64((maxValueTemplate + 1) * (maxValueTemplate + 1) * templateLength)

	parameters := Parameters{int64(maxValueTemplate), maxValueIpfe, 16, templateLength, templateLength + 2}

	clients := make([]Client, numberOfClients)
	referenceTemplates := make([][]int64, numberOfClients)
	for index := range referenceTemplates {
		referenceTemplates[index] = randomVector(int64(maxValueTemplate)+1, templateLength)
	}
	probeTemplates := make([][]int64, numberOfClients)
	for index := range probeTemplates {
		probeTemplates[index] = randomVector(int64(maxValueTemplate)+1, templateLength)
	}
	var network bytes.Buffer
	authenticationResults := make([]*big.Int, numberOfClients)

	// client enroll test
	start := time.Now()
	for index := range clients {
		clients[index].ClientEnrol(parameters, referenceTemplates[index], &network)
	}
	clientEnrolTime := time.Since(start)
	server := NewServer(parameters) // encoded templates have two entries more

	// server enroll test
	start = time.Now()
	for i := 0; i < numberOfClients; i++ {
		server.ServerEnrol(&network)
	}
	serverEnrolTime := time.Since(start)

	// client authentication test
	start = time.Now()
	for index, client := range clients {
		client.ClientAuthentication(probeTemplates[index], &network)
	}
	clientAuthenticationTime := time.Since(start)

	// server authentication test
	start = time.Now()
	for index := range clients {
		ip := server.ServerAuthentication(&network)
		authenticationResults[index] = ip //debug. Uncomment for performance evaluation!
	}
	serverAuthenticationTime := time.Since(start)

	for index, result := range authenticationResults {
		expectedResult := euclidianDistance(referenceTemplates[index], probeTemplates[index])
		if result.Cmp(big.NewInt(expectedResult)) != 0 {
			fmt.Println(referenceTemplates[index])
			fmt.Println(probeTemplates[index])
			log.Fatalf("Wrong result. Should have been %d but was %d", expectedResult, result)
		}
	}

	file, err := os.OpenFile("data/runtime.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	exitOnError(err)
	defer file.Close()

	if ipfeTiming {
		fmt.Println("total ipfeClientEnrollmentTime (nano sec.):", int64(ipfeClientEnrollmentTime)/int64(numberOfClients))
		fmt.Println("total ipfeClientAuthenticationTime (nano sec.):", int64(ipfeClientAuthenticationTime)/int64(numberOfClients))
		fmt.Println("total ipfeServerAuthenticationTime (nano sec.):", int64(ipfeServerAuthenticationTime)/int64(numberOfClients))

		_, err = file.WriteString(fmt.Sprintf("%d;-1;-1;-1;-1;%d;%d;%d\n", templateLength,
			int64(ipfeClientEnrollmentTime)/int64(numberOfClients),
			int64(ipfeClientAuthenticationTime)/int64(numberOfClients),
			int64(ipfeServerAuthenticationTime)/int64(numberOfClients)))
		exitOnError(err)
	} else {
		fmt.Println("total clientEnrolTime (nano sec.):", int64(clientEnrolTime)/int64(numberOfClients))
		fmt.Println("total serverEnrolTime (nano sec.):", int64(serverEnrolTime)/int64(numberOfClients))
		fmt.Println("total clientAuthenticationTime (nano sec.):", int64(clientAuthenticationTime)/int64(numberOfClients))
		fmt.Println("total serverAuthenticationTime (nano sec.):", int64(serverAuthenticationTime)/int64(numberOfClients))

		_, err = file.WriteString(fmt.Sprintf("%d;%d;%d;%d;%d;-1;-1;-1\n", templateLength,
			int64(clientEnrolTime)/int64(numberOfClients),
			int64(serverEnrolTime)/int64(numberOfClients),
			int64(clientAuthenticationTime)/int64(numberOfClients),
			int64(serverAuthenticationTime)/int64(numberOfClients)))
		exitOnError(err)
	}

}

func innerProduct(x, y []int64) int64 {
	if len(x) != len(y) {
		log.Fatalf("Lengths are but %d, %d should be equal", len(x), len(y))
	}
	result := int64(0)
	for index := range x {
		result += x[index] * y[index]
	}
	return result
}

func euclidianDistance(x, y []int64) int64 {
	if len(x) != len(y) {
		log.Fatalf("Lengths are but %d, %d should be equal", len(x), len(y))
	}
	result := int64(0)
	for index := range x {
		result += (x[index] - y[index]) * (x[index] - y[index])
	}
	return result
}
