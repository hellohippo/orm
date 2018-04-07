package main

import (
	"bytes"
	"fmt"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	_ "github.com/mattn/go-sqlite3"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/apex/log"
	"github.com/phogolabs/gom"
	"github.com/phogolabs/gom/example"
	"github.com/phogolabs/gom/example/database/model"
	lk "github.com/ulule/loukoum"
)

func main() {
	script, err := example.Asset("database/script/20180406191137.sql")
	if err != nil {
		log.WithError(err).Fatal("Failed to load embedded resource")
	}

	if err := gom.Load(bytes.NewBuffer(script)); err != nil {
		log.WithError(err).Fatal("Failed to load script")
	}

	gateway, err := gom.Open("sqlite3", "./gom.db")

	if err != nil {
		log.WithError(err).Fatal("Failed to open database connection")
	}

	defer gateway.Close()

	for i := 0; i < 10; i++ {
		var lastName interface{}

		if i%2 == 0 {
			lastName = randomdata.LastName()
		}

		query := lk.Insert("users").
			Set(
				lk.Pair("id", time.Now().UnixNano()),
				lk.Pair("first_name", randomdata.FirstName(randomdata.Male)),
				lk.Pair("last_name", lastName),
			)

		if _, err = gateway.Exec(query); err != nil {
			log.WithError(err).Fatal("Failed to insert new user")
		}
	}

	users := []model.User{}

	if err = gateway.Select(&users, gom.Command("show-users")); err != nil {
		log.WithError(err).Fatal("Failed to select all users")
	}

	validate := validator.New()

	for _, user := range users {
		if err := validate.Struct(user); err != nil {
			log.WithError(err).Error("Failed to validate user")
			continue
		}

		fmt.Printf("User ID: %v\n", user.Id)
		fmt.Printf("First Name: %v\n", user.FirstName)

		if user.LastName.Valid {
			fmt.Printf("Last Name: %v\n", user.LastName.String)
		} else {
			fmt.Println("Last Name: null")
		}

		fmt.Println("---")
	}
}
