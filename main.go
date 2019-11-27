package main

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type User struct {
	Username string `firestore:"username,omitempty"`
	Password string `firestore:"password,omitempty"`
	Nama     string `firestore:"nama,omitempty"`
}

type Todo struct {
	IdTodo    string `firestore:"idTodo,omitempty"`
	Tugas     string `firestore:"tugas,omitempty"`
	Deskripsi string `firestore:"deskripsi,omitempty"`
	Deadline  string `firestore:"deadline,omitempty"`
	Status    bool   `firestore:"status,omitempty"`
}

var ctx context.Context

const userCollection = "users"
const todoCollection = "todo"

func main() {
	ctx = context.Background()

	router := gin.Default()

	userGroup := router.Group("/users")
	{
		userGroup.POST("/", addUsers)
		userGroup.POST("/login", login)
		userGroup.PUT("/:username", updateUser)
	}

	todoGroup := router.Group("/todos")
	{
		todoGroup.GET("/:username", getAllTodos)
		todoGroup.POST("/:username", addTodos)
		todoGroup.PUT("/:username/:idTodo", updateTodos)
		todoGroup.DELETE("/:username/:idTodo", deleteTodos)
	}

	router.Run()
}

func connectFirebase() *firestore.Client {
	sa := option.WithCredentialsFile("serviceAccount.json")

	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	var errClient error
	client, errClient := app.Firestore(ctx)
	if errClient != nil {
		log.Fatalln(errClient)
	}
	return client
}

func getAllTodos(c *gin.Context) {
	username := c.Param("username")

	client := connectFirebase()
	defer client.Close()

	todoIter := client.Collection(userCollection).Doc(username).Collection(todoCollection).Documents(ctx)
	todoSnaps, _ := todoIter.GetAll()

	todos := []Todo{}
	for _, todoSnap := range todoSnaps {
		var todo Todo
		todoSnap.DataTo(&todo)
		todo.IdTodo = todoSnap.Ref.ID
		todos = append(todos, todo)
	}

	c.JSON(http.StatusCreated, gin.H{
		"pesan": "Berhasil mendapatkan semua todo",
		"todo":  todos,
	})
}

func addTodos(c *gin.Context) {
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"pesan": "Gagal menambahkan todo baru, pastikan untuk mengisi semua parameter yang dibutuhkan",
		})
		return
	}
	todo.Status = false

	username := c.Param("username")

	client := connectFirebase()
	defer client.Close()

	client.Collection(userCollection).Doc(username).Collection(todoCollection).Add(ctx, todo)
	c.JSON(http.StatusCreated, gin.H{
		"pesan":     "Berhasil menambahkan todo baru",
		"todo_baru": todo,
	})
}

func updateTodos(c *gin.Context) {
	var updatedTodo Todo
	if err := c.ShouldBindJSON(&updatedTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"pesan": "Gagal update todo, pastikan untuk mengisi semua parameter yang dibutuhkan",
		})
		return
	}

	username := c.Param("username")
	idTodo := c.Param("idTodo")

	client := connectFirebase()
	defer client.Close()

	client.Collection(userCollection).Doc(username).Collection(todoCollection).Doc(idTodo).Set(ctx, updatedTodo)
	c.JSON(http.StatusOK, gin.H{
		"pesan":     "Berhasil melakukan update user",
		"todo_baru": updatedTodo,
	})
}

func deleteTodos(c *gin.Context) {
	username := c.Param("username")
	idTodo := c.Param("idTodo")

	client := connectFirebase()
	defer client.Close()

	client.Collection(userCollection).Doc(username).Collection(todoCollection).Doc(idTodo).Delete(ctx)
	c.JSON(http.StatusOK, gin.H{
		"pesan": "Berhasil melakukan penghapusan todo",
	})
}

func addUsers(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"pesan": "Gagal menambahkan pengguna baru, pastikan untuk mengisi semua parameter yang dibutuhkan",
		})
		return
	}

	client := connectFirebase()
	defer client.Close()

	client.Collection(userCollection).Doc(user.Username).Set(ctx, user)
	c.JSON(http.StatusCreated, gin.H{
		"pesan":         "Berhasil menambahkan pengguna baru",
		"pengguna_baru": user,
	})
}

func login(c *gin.Context) {
	var body gin.H
	c.BindJSON(&body)

	client := connectFirebase()
	defer client.Close()

	var user User
	userSnap, _ := client.Collection(userCollection).Doc(body["username"].(string)).Get(ctx)
	userSnap.DataTo(&user)

	if user.Password != body["password"] {
		c.JSON(http.StatusUnauthorized, gin.H{
			"pesan": "Password yang dimasukkan salah",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pesan":         "Berhasil melakukan login",
		"pengguna_baru": user,
	})
}

func updateUser(c *gin.Context) {
	var updatedUser User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"pesan": "Gagal update pengguna baru, pastikan untuk mengisi semua parameter yang dibutuhkan",
		})
		return
	}

	username := c.Param("username")

	client := connectFirebase()
	defer client.Close()

	client.Collection(userCollection).Doc(username).Set(ctx, updatedUser)

	c.JSON(http.StatusOK, gin.H{
		"pesan":         "Berhasil melakukan update user",
		"pengguna_baru": updatedUser,
	})
}
