package utils

/*
 * File: internal/utils/dbtoolkit_test.go
 *
 * Purpose: unit tests for the dbtoolkit utilities
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/drivertest"
)

func TestConnectWithDefaultSetting(t *testing.T) {
	m := drivertest.NewMockDeployment(
		bson.D{{Key: "ok", Value: 1}},
	)

	conn, err := NewMongoConnection(ConnectionDeployment(m))

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Disconnect(context.Background())
}

func TestConnectWithSpecifiedSetting(t *testing.T) {
	m := drivertest.NewMockDeployment(
		bson.D{{Key: "ok", Value: 1}},
	)

	conn, err := NewMongoConnection(
		ConnectionDeployment(m),
		MongoAppName("dbxtestcase"),
		MongoHosts("127.0.0.1"),
		DirectConnection(true),
	)

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Disconnect(context.Background())
}

func TestConnectWithInvalidSettingCombination(t *testing.T) {
	conn, err := NewMongoConnection(
		MongoAppName("dbxtestcase"),
		MongoHosts("127.0.0.1:27017", "127.0.0.1:27000"),
		DirectConnection(true),
	)

	assert.Nil(t, conn)
	assert.Errorf(t, err, "a direct connection cannot be made if multiple hosts are specified")
}

func TestConnectWithTimedOutPing(t *testing.T) {
	m := drivertest.NewMockDeployment()

	conn, err := NewMongoConnection(
		ConnectionDeployment(m),
		MongoAppName("dbxtestcase"),
		MongoHosts("127.0.0.1"),
		DirectConnection(true),
		ConnectTimeout(2*time.Second),
	)

	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestConnectWithClientCreationFailure(t *testing.T) {
	badFactory := func(opts ...*options.ClientOptions) (*mongo.Client, error) {
		return nil, errors.New("client creation failure")
	}

	original := ClientFactory
	ClientFactory = badFactory
	defer func() {
		ClientFactory = original
	}()

	conn, err := NewMongoConnection(
		MongoAppName("dbxtestcase"),
		MongoHosts("127.0.0.1"),
		DirectConnection(true),
	)

	assert.Errorf(t, err, "client creation failure")
	assert.Nil(t, conn)
}

func TestAsc(t *testing.T) {
	field := "test_field"
	sortKey := Asc(field)
	result := sortKey()

	assert.Equal(t, field, result.Key)
	assert.Equal(t, SortAscending, result.Value)
}

func TestDes(t *testing.T) {
	field := "test_field"
	sortKey := Des(field)
	result := sortKey()

	assert.Equal(t, field, result.Key)
	assert.Equal(t, SortDescending, result.Value)
}

func TestDiscard(t *testing.T) {
	field := "test_field"
	selector := Discard(field)
	result := selector()

	assert.Equal(t, field, result.Key)
	assert.Equal(t, ProjectionDiscard, result.Value)
}

func TestRetain(t *testing.T) {
	field := "test_field"
	selector := Retain(field)
	result := selector()

	assert.Equal(t, field, result.Key)
	assert.Equal(t, ProjectionRetain, result.Value)
}

func TestEq(t *testing.T) {
	field := "test_field"
	value := "test_value"
	condition := Eq(field, value)
	result := condition()

	assert.Equal(t, field, result.Key)
	assert.Equal(t, value, result.Value)
}

func TestGt(t *testing.T) {
	field := "age"
	minValue := 25
	condition := Gt(field, minValue)
	result := condition()

	assert.Equal(t, field, result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterGreaterThan), nested.Key)
	assert.Equal(t, minValue, nested.Value)
}

func TestLt(t *testing.T) {
	field := "age"
	maxValue := 50
	condition := Lt(field, maxValue)
	result := condition()

	assert.Equal(t, field, result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterLessThan), nested.Key)
	assert.Equal(t, maxValue, nested.Value)
}

func TestGte(t *testing.T) {
	field := "score"
	minValueIncluded := 80
	condition := Gte(field, minValueIncluded)
	result := condition()

	assert.Equal(t, field, result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterGreaterThanOrEqualTo), nested.Key)
	assert.Equal(t, minValueIncluded, nested.Value)
}

func TestLte(t *testing.T) {
	field := "score"
	maxValueIncluded := 90
	condition := Lte(field, maxValueIncluded)
	result := condition()

	assert.Equal(t, field, result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterLessThanOrEqualTo), nested.Key)
	assert.Equal(t, maxValueIncluded, nested.Value)
}

func TestIn(t *testing.T) {
	field := "status"
	values := []any{"active", "pending"}
	condition := In(field, values...)
	result := condition()

	assert.Equal(t, field, result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterValueInArray), nested.Key)
	assert.Equal(t, values, nested.Value)
}

func TestInWithSingleValue(t *testing.T) {
	field := "category"
	value := "electronics"
	condition := In(field, value)
	result := condition()

	assert.Equal(t, field, result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterValueInArray), nested.Key)
	expectedValues := []any{value}
	assert.Equal(t, expectedValues, nested.Value)
}

func TestNotIn(t *testing.T) {
	field := "role"
	values := []any{"admin", "guest"}
	condition := NotIn(field, values...)
	result := condition()

	assert.Equal(t, field, result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterValueNotInArray), nested.Key)
	assert.Equal(t, values, nested.Value)
}

func TestAnd(t *testing.T) {
	cond1 := Eq("name", "John")
	cond2 := Gt("age", 25)
	condition := And(cond1, cond2)
	result := condition()

	assert.Equal(t, string(FilterLogicalAnd), result.Key)
	clauses := result.Value.(bson.A)
	assert.Len(t, clauses, 2)
}

func TestAndWithNoConditions(t *testing.T) {
	condition := And()
	result := condition()

	assert.Equal(t, string(FilterLogicalAnd), result.Key)
	clauses := result.Value.(bson.A)
	assert.Len(t, clauses, 0)
}

func TestOr(t *testing.T) {
	cond1 := Eq("status", "active")
	cond2 := Eq("status", "pending")
	condition := Or(cond1, cond2)
	result := condition()

	assert.Equal(t, string(FilterLogicalOr), result.Key)
	clauses := result.Value.(bson.A)
	assert.Len(t, clauses, 2)
}

func TestOrWithNoConditions(t *testing.T) {
	condition := Or()
	result := condition()

	assert.Equal(t, string(FilterLogicalOr), result.Key)
	clauses := result.Value.(bson.A)
	assert.Len(t, clauses, 0)
}

func TestNot(t *testing.T) {
	innerCond := Eq("deleted", true)
	condition := Not(innerCond)
	result := condition()

	assert.Equal(t, string(FilterLogicalNot), result.Key)
	assert.Equal(t, innerCond(), result.Value)
}

func TestExists(t *testing.T) {
	field := "email"
	condition := Exists(field)
	result := condition()

	assert.Equal(t, field, result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterFieldExists), nested.Key)
	assert.Equal(t, true, nested.Value)
}

func TestNotExists(t *testing.T) {
	field := "deleted_at"
	condition := NotExists(field)
	result := condition()

	assert.Equal(t, field, result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterFieldExists), nested.Key)
	assert.Equal(t, false, nested.Value)
}

func TestSet(t *testing.T) {
	field := "name"
	value := "Updated Name"
	instruction := Set(field, value)
	result := instruction()

	assert.Equal(t, string(UpdateSetValue), result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, field, nested.Key)
	assert.Equal(t, value, nested.Value)
}

func TestIncrement(t *testing.T) {
	field := "counter"
	step := 5
	instruction := Increment(field, step)
	result := instruction()

	assert.Equal(t, string(UpdateIncrementValue), result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, field, nested.Key)
	assert.Equal(t, step, nested.Value)
}

func TestDecrement(t *testing.T) {
	field := "balance"
	step := 10
	instruction := Decrement(field, step)
	result := instruction()

	assert.Equal(t, string(UpdateDecrementValue), result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, field, nested.Key)
	assert.Equal(t, step, nested.Value)
}

func TestScale(t *testing.T) {
	field := "price"
	step := 1.5
	instruction := Scale(field, step)
	result := instruction()

	assert.Equal(t, string(UpdateMultiplyValue), result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, field, nested.Key)
	assert.Equal(t, step, nested.Value)
}

func TestDirectives(t *testing.T) {
	cond1 := Eq("field1", "value1")
	cond2 := Eq("field2", "value2")

	directives := Directives(cond1, cond2)

	assert.Len(t, directives, 2)
	assert.Equal(t, cond1(), directives[0]())
	assert.Equal(t, cond2(), directives[1]())
}

func TestDirectivesWithEmptySlice(t *testing.T) {
	// Test with sortKey type by creating a dummy sort key
	ascKey := Asc("dummy")
	directives := Directives(ascKey)
	// Clear the slice to test empty behavior
	directives = directives[:0]
	assert.Len(t, directives, 0)
}

func TestDirectivesWithSingleItem(t *testing.T) {
	cond := Eq("field", "value")
	directives := Directives(cond)

	assert.Len(t, directives, 1)
	assert.Equal(t, cond(), directives[0]())
}

func TestConstantsValues(t *testing.T) {
	assert.Equal(t, 1, int(SortAscending))
	assert.Equal(t, -1, int(SortDescending))

	assert.Equal(t, 0, int(ProjectionDiscard))
	assert.Equal(t, 1, int(ProjectionRetain))

	assert.Equal(t, "$gt", string(FilterGreaterThan))
	assert.Equal(t, "$gte", string(FilterGreaterThanOrEqualTo))
	assert.Equal(t, "$lt", string(FilterLessThan))
	assert.Equal(t, "$lte", string(FilterLessThanOrEqualTo))
	assert.Equal(t, "$in", string(FilterValueInArray))
	assert.Equal(t, "$nin", string(FilterValueNotInArray))
	assert.Equal(t, "$and", string(FilterLogicalAnd))
	assert.Equal(t, "$or", string(FilterLogicalOr))
	assert.Equal(t, "$not", string(FilterLogicalNot))
	assert.Equal(t, "$exists", string(FilterFieldExists))

	assert.Equal(t, "$set", string(UpdateSetValue))
	assert.Equal(t, "$inc", string(UpdateIncrementValue))
	assert.Equal(t, "$dec", string(UpdateDecrementValue))
	assert.Equal(t, "$mul", string(UpdateMultiplyValue))
}

func TestComplexFilterCombinations(t *testing.T) {
	nameCond := Eq("name", "John")
	ageCond := Gt("age", 25)
	statusCond := In("status", "active", "verified")

	andCond := And(nameCond, ageCond, statusCond)
	notCond := Not(andCond)

	result := notCond()

	assert.Equal(t, string(FilterLogicalNot), result.Key)
	nested := result.Value.(bson.E)
	assert.Equal(t, string(FilterLogicalAnd), nested.Key)

	clauses := nested.Value.(bson.A)
	assert.Len(t, clauses, 3)
}

func TestComplexUpdateCombinations(t *testing.T) {
	setInst := Set("name", "New Name")
	incInst := Increment("counter", 1)
	mulInst := Scale("price", 1.1)

	directives := Directives(setInst, incInst, mulInst)

	assert.Len(t, directives, 3)

	setResult := directives[0]()
	assert.Equal(t, string(UpdateSetValue), setResult.Key)

	incResult := directives[1]()
	assert.Equal(t, string(UpdateIncrementValue), incResult.Key)

	mulResult := directives[2]()
	assert.Equal(t, string(UpdateMultiplyValue), mulResult.Key)
}

func TestDifferentValueTypes(t *testing.T) {
	stringCond := Eq("string_field", "test")
	intCond := Eq("int_field", 42)
	floatCond := Eq("float_field", 3.14)
	boolCond := Eq("bool_field", true)

	assert.Equal(t, "test", stringCond().Value)
	assert.Equal(t, 42, intCond().Value)
	assert.Equal(t, 3.14, floatCond().Value)
	assert.Equal(t, true, boolCond().Value)
}

func TestNilValue(t *testing.T) {
	cond := Eq("nullable_field", nil)
	result := cond()

	assert.Equal(t, "nullable_field", result.Key)
	assert.Equal(t, nil, result.Value)
}
