package mongodb-activity

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TEST_URI  = "mongodb://localhost:27017"
	TEST_DB   = "test"
	TEST_COLL = "items"
)

var coll *mongo.Collection

func init() {
	//todo implement shared sessions
	// client, err := mongo.NewClient(TEST_URI)

	//To remove below  error:
	// data not inserted topology is closed

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(TEST_URI))
	if err != nil {
		// warn and skip tests
	}

	db := client.Database(TEST_DB)
	coll = db.Collection(TEST_COLL)
}

var activityMetadata *activity.Metadata

func getActivityMetadata() *activity.Metadata {
	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}

		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
	}

	return activityMetadata
}

func randomString(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func insert(dataVal interface{}) (interface{}, error) {

	result, err := coll.InsertOne(
		context.Background(),
		dataVal,
	)
	if err != nil {
		return nil, err
	}

	logger.Debug("Inserted: ", result.InsertedID)

	return result.InsertedID, nil
}

func delete(id interface{}) {
	oid := id.(primitive.ObjectID)
	_, err := coll.DeleteOne(context.Background(), bson.M{"_id": oid})
	if err != nil {
		logger.Debugf("Error Deleting [%s] : %s", id, err.Error())
		return
	}
	logger.Debug("Deleted", id)
}

// TestDelete
func TestDelete(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	name := randomString(5)
	val := map[string]interface{}{"name": name, "value": "blah"}
	_, err := insert(val)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	tc.SetInput("uri", TEST_URI)
	tc.SetInput("dbName", TEST_DB)
	tc.SetInput("collection", TEST_COLL)
	tc.SetInput("method", `DELETE`)

	tc.SetInput(ivKeyName, "name")
	tc.SetInput(ivKeyValue, name)

	_, deleteErr := act.Eval(tc)
	if deleteErr != nil {
		t.Error("data not deleted")
		t.Fail()
	}
}

// TestInsert
func TestInsert(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	tc.SetInput("uri", TEST_URI)
	tc.SetInput("dbName", TEST_DB)
	tc.SetInput("collection", TEST_COLL)
	tc.SetInput("method", `INSERT`)

	name := randomString(5)
	val := map[string]interface{}{"name": name, "value1": "foo", "value2": "foo2"}
	tc.SetInput(ivData, val)
	//tc.SetInput(ivKeyName, "key")
	//tc.SetInput(ivKeyValue, "value")

	_, insertErr := act.Eval(tc)
	if insertErr != nil {
		t.Error("data not inserted", insertErr)
		t.Fail()
	}
	fmt.Println("Output ", reflect.TypeOf(tc.GetOutput(ovOutput)))

	delete(tc.GetOutput(ovOutput))
}

// TestReplace
func TestReplace(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	name := randomString(5)
	val := map[string]interface{}{"name": name, "value1": "foo", "value2": "foo2"}
	id, err := insert(val)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	tc.SetInput("uri", TEST_URI)
	tc.SetInput("dbName", TEST_DB)
	tc.SetInput("collection", TEST_COLL)
	tc.SetInput("method", `REPLACE`)

	val2 := map[string]interface{}{"name": name, "value1": "bar1", "value2": "bar2"}

	tc.SetInput(ivKeyName, "name")
	tc.SetInput(ivKeyValue, name)
	tc.SetInput(ivData, val2)

	_, replaceErr := act.Eval(tc)
	if replaceErr != nil {
		t.Error("data not replaced", replaceErr)
		t.Fail()
	}

	delete(id)
}

// TestReplace
func TestUpdate(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	name := randomString(5)
	val := map[string]interface{}{"name": name, "value1": "foo", "value2": "foo2"}
	id, err := insert(val)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	tc.SetInput("uri", TEST_URI)
	tc.SetInput("dbName", TEST_DB)
	tc.SetInput("collection", TEST_COLL)
	tc.SetInput("method", `UPDATE`)

	val2 := map[string]interface{}{"name": name, "value1": "bar1"}

	tc.SetInput(ivKeyName, "name")
	tc.SetInput(ivKeyValue, name)
	tc.SetInput(ivData, val2)

	_, updateErr := act.Eval(tc)
	if updateErr != nil {
		t.Error("update error", updateErr)
		t.Fail()
	}

	delete(id)
}

// TestGet
func TestGet(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	name := randomString(5)
	val := map[string]interface{}{"name": name, "value1": "foo", "value2": "foo2"}
	id, err := insert(val)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	tc.SetInput("uri", TEST_URI)
	tc.SetInput("dbName", TEST_DB)
	tc.SetInput("collection", TEST_COLL)
	tc.SetInput("method", `GET`)

	tc.SetInput(ivKeyName, "name")
	tc.SetInput(ivKeyValue, name)

	_, getErr := act.Eval(tc)
	if getErr != nil {
		t.Error("unable to get data", getErr)
		t.Fail()
	}

	result := tc.GetOutput(ovOutput)
	assert.NotNil(t, result)

	delete(id)
}
