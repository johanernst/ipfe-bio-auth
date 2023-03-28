package main

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	mrand "math/rand"
	"time"

	"golang.org/x/crypto/nacl/sign"
	"golang.org/x/crypto/sha3"

	"github.com/fentec-project/gofe/data"
	"github.com/fentec-project/gofe/innerprod/fullysec"
)

var ipfeTiming = false

var ipfeClientEnrollmentTime time.Duration
var ipfeClientAuthenticationTime time.Duration
var ipfeServerAuthenticationTime time.Duration
var start time.Time

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Parameters struct {
	maxValueTemplate  int64
	maxValueIpfe      int64
	securityParameter int
	templateLength    int
	ipfeVectorLength  int
}

type Client struct {
	rid                []byte
	signatureSecretKey *[64]byte
	scheme             *fullysec.FHIPE
	masterKey          *fullysec.FHIPESecKey
	parameters         Parameters
}

type Server struct {
	database   map[string]dbItem
	scheme     *fullysec.FHIPE
	parameters Parameters
}

type dbItem struct {
	signaturePublicKey *[32]byte
	decryptionKey      *fullysec.FHIPEDerivedKey
}

type enrollmentMessage struct {
	Rid                []byte
	SignaturePublicKey *[32]byte
	DecryptionKey      *fullysec.FHIPEDerivedKey
}

type authenticationMessage struct {
	Rid        []byte
	Ciphertext *fullysec.FHIPECipher
}

// helper function that converts a []byte rid to a hexstring
func ridToString(rid []byte) string {
	return hex.EncodeToString(rid)
}

// ClientEnrol: x
func (client *Client) ClientEnrol(parameters Parameters, referenceTemplate []int64, network *bytes.Buffer) {
	if parameters.templateLength != len(referenceTemplate) {
		log.Fatalf("Length of referenceTemplate (%d) does not match parameters.templateLength (%d)",
			referenceTemplate, parameters.templateLength)
	}

	// encode the reference templates so that it can be used for computing the euclidian distance
	encodedTemplate := encodeEuclidian(referenceTemplate, true)
	// fmt.Printf("%+v\n", encodedTemplate)

	// create random rid for the client
	client.rid = make([]byte, parameters.securityParameter)
	_, err := rand.Read(client.rid)
	exitOnError(err)

	// create signature keypair
	pk, sk, err := sign.GenerateKey(rand.Reader)
	exitOnError(err)
	client.signatureSecretKey = sk

	// create ipfe master key, and functional decryption key for the server
	if ipfeTiming {
		start = time.Now()
	}
	client.scheme, err = fullysec.NewFHIPE(len(encodedTemplate),
		big.NewInt(parameters.maxValueIpfe), big.NewInt(parameters.maxValueIpfe))
	exitOnError(err)
	client.masterKey, err = client.scheme.GenerateMasterKey()
	exitOnError(err)
	decryptionKey, err := client.scheme.DeriveKey(encodedTemplate, client.masterKey)
	exitOnError(err)
	if ipfeTiming {
		ipfeClientEnrollmentTime += time.Since(start)
	}

	message := enrollmentMessage{client.rid, pk, decryptionKey}

	encoder := gob.NewEncoder(network)
	err = encoder.Encode(&message)
	exitOnError(err)
}

func (client *Client) ClientAuthentication(probeTemplate []int64, network *bytes.Buffer) {

	// encode the reference templates so that it can be used for computing the euclidian distance
	encodedTemplate := encodeEuclidian(probeTemplate, false)

	// encrypt the encoded template
	if ipfeTiming {
		start = time.Now()
	}
	ciphertext, err := client.scheme.Encrypt(encodedTemplate, client.masterKey)
	exitOnError(err)
	if ipfeTiming {
		ipfeClientAuthenticationTime += time.Since(start)
	}

	// hash and sign the message
	message := authenticationMessage{client.rid, ciphertext}
	var encodedCiphertext bytes.Buffer
	hashEncoder := gob.NewEncoder(&encodedCiphertext)
	err = hashEncoder.Encode(message)
	exitOnError(err)

	hashedCiphertext := make([]byte, 64)
	sha3.ShakeSum256(hashedCiphertext, encodedCiphertext.Bytes())

	signature := sign.Sign(nil, hashedCiphertext, client.signatureSecretKey)

	// send the message at the signature
	networkEncoder := gob.NewEncoder(network)
	err = networkEncoder.Encode(&message)
	exitOnError(err)
	err = networkEncoder.Encode(&signature)
	exitOnError(err)
}

func NewServer(parameters Parameters) *Server {
	scheme, err := fullysec.NewFHIPE(parameters.ipfeVectorLength, big.NewInt(parameters.maxValueIpfe),
		big.NewInt(parameters.maxValueIpfe))
	exitOnError(err)
	return &Server{make(map[string]dbItem), scheme, parameters}
}

func (server *Server) ServerEnrol(network *bytes.Buffer) {
	decoder := gob.NewDecoder(network)
	var decodedMessage enrollmentMessage
	err := decoder.Decode(&decodedMessage)
	exitOnError(err)

	mapItem := dbItem{decodedMessage.SignaturePublicKey, decodedMessage.DecryptionKey}
	server.database[ridToString(decodedMessage.Rid)] = mapItem
}

// TODO I could reduce the interval in which the algoritm searches for the DLOG,
// because the maximum euclidean distance is smaller than the maximum inner-
// product, because the encoded template contains the L2 norm as entry.
// ---> I could even further reduce this interval to [0,tau], because we reject
// anyways, if the distance is larger
func (server *Server) ServerAuthentication(network *bytes.Buffer) *big.Int {
	// decode the message
	decoder := gob.NewDecoder(network)
	var decodedMessage authenticationMessage
	err := decoder.Decode(&decodedMessage)
	exitOnError(err)

	var signature []byte
	err = decoder.Decode(&signature)
	exitOnError(err)

	item, exists := server.database[ridToString(decodedMessage.Rid)]
	if !exists {
		errorString := fmt.Sprintf("The following rid is trying to authenticate but is not enrolled: %v", decodedMessage.Rid)
		panic(errorString)
	}

	// verify signature
	var encodedCiphertext bytes.Buffer
	hashEncoder := gob.NewEncoder(&encodedCiphertext)
	err = hashEncoder.Encode(decodedMessage)
	exitOnError(err)

	hashedCiphertext := make([]byte, 64)
	sha3.ShakeSum256(hashedCiphertext, encodedCiphertext.Bytes())
	message, valid := sign.Open(nil, signature, item.signaturePublicKey)
	if !valid {
		log.Fatal("signature invalid\n")
	}
	if !bytes.Equal(message, hashedCiphertext) {
		log.Fatal("signature valid for wrong message\n")
	}

	// compute distance
	if ipfeTiming {
		start = time.Now()
	}
	ip, err := server.scheme.Decrypt(decodedMessage.Ciphertext, item.decryptionKey)
	exitOnError(err)
	if ipfeTiming {
		ipfeServerAuthenticationTime += time.Since(start)
	}
	return ip
}

func encodeEuclidian(template []int64, enrol bool) data.Vector {
	result := data.NewConstantVector(len(template)+2, big.NewInt(0))
	n := len(result)
	multiplier := big.NewInt(1)
	squaredEuclidean := big.NewInt(0)
	if !enrol {
		multiplier = big.NewInt(-2) //probe template has a factor of -2 in every component
	}
	for i := 0; i < len(template); i++ {
		//the +1 is to ensure that all elements are > 0 i.e. invertible
		result[i] = big.NewInt(template[i] + 1)

		squaredEuclidean.Add(squaredEuclidean, new(big.Int).Mul(result[i], result[i]))
		result[i].Mul(result[i], multiplier)
	}

	if enrol { //encode reference template
		result[n-2] = big.NewInt(1)
		result[n-1] = squaredEuclidean
	} else { //encode probe template
		result[n-2] = squaredEuclidean
		result[n-1] = big.NewInt(1)
	}
	return result
}

func randomVector(upperBound int64, length int) []int64 {
	vector := make([]int64, length)
	for index := range vector {
		vector[index] = mrand.Int63n(upperBound)
	}
	return vector
}

func main() {

}
