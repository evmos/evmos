package eip712_test

import (
	"fmt"
	"strings"

	rand "github.com/tendermint/tendermint/libs/rand"

	"github.com/evmos/evmos/v11/ethereum/eip712"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// TestRandomPayloadFlattening generates many random payloads with different JSON values to ensure
// that Flattening works across all inputs.
// Note that this is a fuzz test, although it doesn't use Go's Fuzz testing suite, since there are
// variable input sizes, types, and fields. While it may be possible to translate a single input into
// a JSON object, it would require difficult parsing, and ultimately approximates our randomized unit
// tests as they are.
func (suite *EIP712TestSuite) TestRandomPayloadFlattening() {
	// Re-seed rand generator
	rand.Seed(rand.Int64())

	numTestObjects := 15
	for i := 0; i < numTestObjects; i++ {
		suite.Run(fmt.Sprintf("Flatten%d", i), func() {
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
		m := suite.createRandomJSONObject()
		msgs[i] = m
	}

	payload, err := sjson.Set(payload, "msgs", msgs)
	suite.Require().NoError(err)

	return gjson.Parse(payload)
}

// createRandomJSONObject creates a JSON object with random fields.
func (suite *EIP712TestSuite) createRandomJSONObject() gjson.Result {
	var err error
	payloadRaw := ""

	numFields := suite.randomInRange(0, 16)
	for i := 0; i < numFields; i++ {
		key := suite.generateRandomString(12, 36)

		randField := suite.createRandomJSONField(i, 0)
		payloadRaw, err = sjson.Set(payloadRaw, key, randField)
		suite.Require().NoError(err)
	}

	return gjson.Parse(payloadRaw)
}

// createRandomJSONField creates a random field with a random JSON type, with the possibility of
// nested fields up to depth.
func (suite *EIP712TestSuite) createRandomJSONField(t int, depth int) interface{} {
	constNumTypes := 5

	switch t % constNumTypes {
	case 0:
		// Rand bool
		return rand.Intn(2) == 0
	case 1:
		// Rand string
		return suite.generateRandomString(10, 48)
	case 2:
		// Rand num
		return (rand.Float64() - 0.5) * 100000000000
	case 3, 4:
		// Rand array (3) or object (4)
		arr := make([]interface{}, rand.Intn(10))
		obj := make(map[string]interface{})

		for i := range arr {
			fieldType := rand.Intn(constNumTypes)
			if depth == constNumTypes {
				// Max depth
				fieldType = rand.Intn(constNumTypes - 2)
			}

			randField := suite.createRandomJSONField(fieldType, depth+1)

			if t%constNumTypes == 3 {
				arr[i] = randField
			} else {
				obj[suite.generateRandomString(10, 48)] = randField
			}
		}

		if t%constNumTypes == 3 {
			return arr
		}
		return obj
	default:
		// Null
		return nil
	}
}

// generateRandomString generates a random string with the given properties.
func (suite *EIP712TestSuite) generateRandomString(minLength int, maxLength int) string {
	bzLen := suite.randomInRange(minLength, maxLength)
	bz := make([]byte, bzLen)

	for i := 0; i < bzLen; i++ {
		bz[i] = byte(suite.randomInRange(65, 127))
	}

	str := string(bz)
	// Remove control characters, since they will make JSON invalid
	str = strings.ReplaceAll(str, "{", "")
	str = strings.ReplaceAll(str, "}", "")
	str = strings.ReplaceAll(str, "]", "")
	str = strings.ReplaceAll(str, "[", "")

	return str
}

// randomInRange provides a random integer between [min, max)
func (suite *EIP712TestSuite) randomInRange(min int, max int) int {
	return rand.Intn(max-min) + min
}
