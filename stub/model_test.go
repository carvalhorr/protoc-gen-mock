package stub

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJsonString_Matches_TwoEqualJsonStrings(t *testing.T) {
	str1 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2}}")
	str2 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2}}")
	assert.True(t, str1.Matches(str2))
}

func TestJsonString_Matches_LeftJsonStringContainsAllFieldsOfRightJsonString(t *testing.T) {
	str1 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2}}")
	str2 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2, \"subfield3\": 4}}")
	assert.True(t, str1.Matches(str2))
}

func TestJsonString_Matches_RightJsonStringContainsAllFieldsOfLeftJsonString(t *testing.T) {
	str1 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2, \"subfield3\": 4}}")
	str2 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2}}")
	assert.False(t, str1.Matches(str2))
}

func TestJsonString_Equals_TwoEqualJsonStrings(t *testing.T) {
	str1 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2}}")
	str2 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2}}")
	assert.True(t, str1.Equals(str2))
}

func TestJsonString_Equals_LeftJsonStringContainsAllFieldsOfRightJsonString(t *testing.T) {
	str1 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2}}")
	str2 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2, \"subfield3\": 4}}")
	assert.False(t, str1.Equals(str2))
}

func TestJsonString_Equals_RightJsonStringContainsAllFieldsOfLeftJsonString(t *testing.T) {
	str1 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2, \"subfield3\": 4}}")
	str2 := JsonString("{\"field1\":{\"subfieldd1\":\"value1\", \"subfield2\": 2}}")
	assert.False(t, str1.Equals(str2))
}
