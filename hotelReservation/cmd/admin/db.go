package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"log"
)

type Admin struct {
	Name     string   `bason:"name"`
	Email    string   `bason:"email"`
	Password string   `bson:"password"`
	Hotels   []string `bason: "hotels"`
	Id       string   `bason: "id"`
}

func initializeDatabase(url string) *mgo.Session {
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	// defer session.Close()
	c := session.DB("admin-db").C("admin")

	name := "Shuangquan"
	email := "test@qq"
	password := "123"
	l := []string{"1", "2", "3", "4"}
	id := "0"
	count, err := c.Find(&bson.M{"name": name}).Count()
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		err = c.Insert(&Admin{
			name,
			email,
			password,
			l,
			id,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	err = c.EnsureIndexKey("name")
	if err != nil {
		log.Fatal(err)
	}
	return session
}
