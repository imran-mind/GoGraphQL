package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
	Task string `json:"task"`
}

var TodoList []Todo
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func init() {
	todo1 := Todo{ID: "a", Text: "A todo not to forget", Done: false}
	todo2 := Todo{ID: "b", Text: "This is the most important", Done: false}
	todo3 := Todo{ID: "c", Text: "Please do this or else", Done: false}
	TodoList = append(TodoList, todo1, todo2, todo3)

	rand.Seed(time.Now().UnixNano())
}

func main() {

	// fmt.Println("============> helloWorld ", HelloWorld())
	// define custom GraphQL ObjectType `todoType` for our Golang struct `Todo`
	// Note that
	// - the fields in our todoType maps with the json tags for the fields in our struct
	// - the field type matches the field type in our struct
	todoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Todo",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"text": &graphql.Field{
				Type: graphql.String,
			},
			"done": &graphql.Field{
				Type: graphql.Boolean,
			},
			"task": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// root mutation
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"createTodo": &graphql.Field{
				Type: todoType, // the return type for this field
				Args: graphql.FieldConfigArgument{
					"text": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"task": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {

					// marshall and cast the argument value
					text, _ := params.Args["text"].(string)
					task, _ := params.Args["task"].(string)
					// perform mutation operation here
					// for e.g. create a Todo and save to DB.

					newTodo := Todo{
						ID:   "id0001",
						Text: text,
						Done: true,
						Task: task,
					}
					fmt.Println("------------------> ", newTodo)
					// return the new Todo object that we supposedly save to DB
					// Note here that
					// - we are returning a `Todo` struct instance here
					// - we previously specified the return Type to be `todoType`
					// - `Todo` struct maps to `todoType`, as defined in `todoType` ObjectConfig`
					// TodoList = append(TodoList, newTodo)
					TodoList = append(TodoList, newTodo)
					return TodoList, nil
				},
			},

			//update opration of TODO
			"updateTodo": &graphql.Field{
				Type:        todoType, // the return type for this field
				Description: "Update existing todo, mark it done or not done",
				Args: graphql.FieldConfigArgument{
					"done": &graphql.ArgumentConfig{
						Type: graphql.Boolean,
					},
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					// marshall and cast the argument value
					done, _ := params.Args["done"].(bool)
					id, _ := params.Args["id"].(string)
					affectedTodo := Todo{}

					// Search list for todo with id and change the done variable
					for i := 0; i < len(TodoList); i++ {
						if TodoList[i].ID == id {
							TodoList[i].Done = done
							// Assign updated todo so we can return it
							affectedTodo = TodoList[i]
							break
						}
					}
					// Return affected todo
					return affectedTodo, nil
				},
			},
		},
	})

	// root query
	// we just define a trivial example here, since root query is required.
	// Test with curl
	// curl -g 'http://localhost:8080/graphql?query={lastTodo{id,text,done}}'
	var rootQuery = graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{

			/*
			   curl -g 'http://localhost:8080/graphql?query={todo(id:"b"){id,text,done}}'
			*/
			"todo": &graphql.Field{
				Type:        todoType,
				Description: "Get single todo",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {

					idQuery, isOK := params.Args["id"].(string)
					if isOK {
						// Search for el with id
						for _, todo := range TodoList {
							if todo.ID == idQuery {
								return todo, nil
							}
						}
					}

					return Todo{}, nil
				},
			},

			"lastTodo": &graphql.Field{
				Type:        todoType,
				Description: "Last todo added",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return TodoList[len(TodoList)-1], nil
				},
			},

			/*
			   curl -g 'http://localhost:8080/graphql?query={todoList{id,text,done}}'
			*/
			"todoList": &graphql.Field{
				Type:        graphql.NewList(todoType),
				Description: "List of todos",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return TodoList, nil
				},
			},
		},
	})

	// define schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})

	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// serve HTTP
	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
	fmt.Println("Now server is running on port 8080")

	// How to make a HTTP request using cUrl
	// -------------------------------------
	// In `graphql-go-handler`, based on the GET/POST and the Content-Type header, it expects the input params differently.
	// This behaviour was ported from `express-graphql`.
	//
	//
	// 1) using GET
	// $ curl -g -GET 'http://localhost:8080/graphql?query=mutation+M{newTodo:createTodo(text:"This+is+a+todo+mutation+example"){text+done}}'
	//
	// 2) using POST + Content-Type: application/graphql
	// $ curl -XPOST http://localhost:8080/graphql -H 'Content-Type: application/graphql' -d 'mutation M { newTodo: createTodo(text: "This is a todo mutation example") { text done } }'
	//
	// 3) using POST + Content-Type: application/json
	// $ curl -XPOST http://localhost:8080/graphql -H 'Content-Type: application/json' -d '{"query": "mutation M { newTodo: createTodo(text: \"This is a todo mutation example\") { text done } }"}'
	//
	// Any of the above would return the following output:
	// {
	//   "data": {
	// 	   "newTodo": {
	// 	     "done": false,
	// 	     "text": "This is a todo mutation example"
	// 	   }
	//   }
	// }
}
