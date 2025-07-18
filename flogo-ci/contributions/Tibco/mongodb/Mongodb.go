package mongodbcommon

import (
	"reflect"
)

type (
	Connection struct {
		Name          string `md:"Name,required"`
		Description   string `md:"Description"`
		ConnectionURI string `md:"ConnectionURI,required"`
		Database      string `md:"Database,required"`
		DocsMetadata  string `md:"DocsMetadata"`
	}
)

//var log = logger.GetLogger("Mongodb")

var cachedConnection map[string]*Connection

func init() {
	cachedConnection = map[string]*Connection{}
}

// GetConnection returns a deserialized Riak conneciton object it does not establish a
// connection with the Riak. If a connection with the same id as in the context is
// present in the cache, that connection from the cache is returned
// func GetConnection(connector interface{}) (connection *Connection, err error) {

// 	connectionObject := connector.(map[string]interface{})
// 	settings := connectionObject["settings"].([]interface{})
// 	id := connectionObject["id"].(string)
// 	connection = cachedConnection[id]
// 	if connection != nil {
// 		//log.Info("Returning cached connection")
// 		return connection, nil
// 	}

// 	connection = &Connection{}
// 	connection.read(settings)

// 	opts := options.Client() //Added this to point to latest mongodb driver

// 	if connection.MongoClient == nil {
// 		client, err := mongo.NewClient(opts.ApplyURI(connection.ConnectionURI))
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
// 		defer cancel()
// 		err = client.Connect(ctx)
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		err = client.Ping(ctx, nil)
// 		if err != nil {
// 			fmt.Println(err)
// 		} else {
// 			//log.Info("Ping success")
// 			connection.MongoClient = client
// 		}
// 	}

// 	cachedConnection[id] = connection
// 	//log.Info("Returning New connection")
// 	return connection, nil

// }
func (k *Connection) Type() string {
	return "Mongo"
}

func (k *Connection) GetConnection() interface{} {
	return k
}
func (k *Connection) ReleaseConnection(connection interface{}) {

}
func (connection *Connection) read(settings interface{}) (err error) {

	connectionRef := reflect.ValueOf(connection).Elem()

	//TBD process other types later, right now only strings
	for _, value := range settings.([]interface{}) {
		element := value.(map[string]interface{})
		field := connectionRef.FieldByName(element["name"].(string))
		// var fieldValue string

		switch element["value"].(type) {

		case string:
			fieldValue := element["value"].(string)
			field.SetString(fieldValue)
		case int64:
			fieldValue := element["value"].(int64)
			field.SetInt(fieldValue)
		case float64:
			fieldValue := int64(element["value"].(float64))
			field.SetInt(fieldValue)
		}

	}

	return
}
