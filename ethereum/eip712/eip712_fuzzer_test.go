package eip712_test

import (
	"fmt"
	"strings"

	rand "github.com/tendermint/tendermint/libs/rand"

	"github.com/evmos/evmos/v11/ethereum/eip712"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type EIP712FuzzTestParams struct {
	numTestObjects        int
	maxNumFieldsPerObject int
	minStringLength       int
	maxStringLength       int
	randomFloatRange      float64
	maxArrayLength        int
	maxObjectDepth        int
}

const (
	NUM_PRIMITIVE_JSON_TYPES = 3
	NUM_JSON_TYPES           = 5
	ASCII_RANGE_START        = 65
	ASCII_RANGE_END          = 127
	TEST_NAME                = "Flatten"
	MSGS_FIELD_NAME          = "msgs"
)

const (
	JSON_BOOL_TYPE   = iota
	JSON_STRING_TYPE = iota
	JSON_FLOAT_TYPE  = iota
	JSON_ARRAY_TYPE  = iota
	JSON_OBJECT_TYPE = iota
)

var params EIP712FuzzTestParams = EIP712FuzzTestParams{
	numTestObjects:        16,
	maxNumFieldsPerObject: 16,
	minStringLength:       16,
	maxStringLength:       48,
	randomFloatRange:      120000000,
	maxArrayLength:        8,
	maxObjectDepth:        4,
}

// TestRandomPayloadFlattening generates many random payloads with different JSON values to ensure
// that Flattening works across all inputs.
// Note that this is a fuzz test, although it doesn't use Go's Fuzz testing suite, since there are
// variable input sizes, types, and fields. While it may be possible to translate a single input into
// a JSON object, it would require difficult parsing, and ultimately approximates our randomized unit
// tests as they are.
func (suite *EIP712TestSuite) TestRandomPayloadFlattening() {
	// Re-seed rand generator
	rand.Seed(rand.Int64())

	for i := 0; i < params.numTestObjects; i++ {
		suite.Run(fmt.Sprintf("%v%d", TEST_NAME, i), func() {
			payload := suite.generateRandomPayload(i)

			flattened, numMessages, err := eip712.FlattenPayloadMessages(payload)

			suite.Require().NoError(err)
			suite.Require().Equal(numMessages, i)

			suite.verifyPayloadAgainstFlattened(payload, flattened)
		})
	}
}

// generateRandomPayload creates a random payload of the desired format, with random sub-objects.
func (suite *EIP712TestSuite) generateRandomPayload(numMessages int) gjson.Result {
	payload := suite.createRandomJSONObject().Raw
	msgs := make([]gjson.Result, numMessages)

	for i := 0; i < numMessages; i++ {
		msgs[i] = suite.createRandomJSONObject()
	}

	payload, err := sjson.Set(payload, MSGS_FIELD_NAME, msgs)
	suite.Require().NoError(err)

	return gjson.Parse(payload)
}

// createRandomJSONObject creates a JSON object with random fields.
func (suite *EIP712TestSuite) createRandomJSONObject() gjson.Result {
	var err error
	payloadRaw := ""

	numFields := suite.createRandomIntInRange(0, params.maxNumFieldsPerObject)
	for i := 0; i < numFields; i++ {
		key := suite.createRandomString()

		randField := suite.createRandomJSONField(i, 0)
		payloadRaw, err = sjson.Set(payloadRaw, key, randField)
		suite.Require().NoError(err)
	}

	return gjson.Parse(payloadRaw)
}

// createRandomJSONField creates a random field with a random JSON type, with the possibility of
// nested fields up to depth objects.
func (suite *EIP712TestSuite) createRandomJSONField(t int, depth int) interface{} {
	switch t % NUM_JSON_TYPES {
	case JSON_BOOL_TYPE:
		return suite.createRandomBoolean()
	case JSON_STRING_TYPE:
		return suite.createRandomString()
	case JSON_FLOAT_TYPE:
		return suite.createRandomFloat()
	case JSON_ARRAY_TYPE:
		return suite.createRandomJSONNestedArray(depth)
	case JSON_OBJECT_TYPE:
		return suite.createRandomJSONNestedObject(depth)
	default:
		return nil
	}
}

// createRandomJSONNestedArray creates an array of random nested JSON fields.
func (suite *EIP712TestSuite) createRandomJSONNestedArray(depth int) []interface{} {
	arr := make([]interface{}, rand.Intn(params.maxArrayLength))
	for i := range arr {
		arr[i] = suite.createRandomJSONNestedField(depth)
	}

	return arr
}

// createRandomJSONNestedObject creates a key-value set of objects with random nested JSON fields.
func (suite *EIP712TestSuite) createRandomJSONNestedObject(depth int) interface{} {
	numFields := rand.Intn(params.maxNumFieldsPerObject)
	obj := make(map[string]interface{})

	for i := 0; i < numFields; i++ {
		subField := suite.createRandomJSONNestedField(depth)

		obj[suite.createRandomString()] = subField
	}

	return obj
}

// createRandomJSONNestedField serves as a helper for createRandomJSONField and returns a random
// subfield to populate an array or object type.
func (suite *EIP712TestSuite) createRandomJSONNestedField(depth int) interface{} {
	var newFieldType int

	if depth == params.maxObjectDepth {
		newFieldType = rand.Intn(NUM_PRIMITIVE_JSON_TYPES)
	} else {
		newFieldType = rand.Intn(NUM_JSON_TYPES)
	}

	return suite.createRandomJSONField(newFieldType, depth+1)
}

func (suite *EIP712TestSuite) createRandomBoolean() bool {
	return rand.Intn(2) == 0
}

func (suite *EIP712TestSuite) createRandomFloat() float64 {
	return (rand.Float64() - 0.5) * params.randomFloatRange
}

func (suite *EIP712TestSuite) createRandomString() string {
	bzLen := suite.createRandomIntInRange(params.minStringLength, params.maxStringLength)
	bz := make([]byte, bzLen)

	for i := 0; i < bzLen; i++ {
		bz[i] = byte(suite.createRandomIntInRange(ASCII_RANGE_START, ASCII_RANGE_END))
	}

	str := string(bz)

	// Remove control characters, since they will make JSON invalid
	str = strings.ReplaceAll(str, "{", "")
	str = strings.ReplaceAll(str, "}", "")
	str = strings.ReplaceAll(str, "]", "")
	str = strings.ReplaceAll(str, "[", "")

	return str
}

// createRandomIntInRange provides a random integer between [min, max)
func (suite *EIP712TestSuite) createRandomIntInRange(min int, max int) int {
	return rand.Intn(max-min) + min
}
