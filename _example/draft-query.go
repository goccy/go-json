package main

import (
	"fmt"

	"github.com/goccy/go-json"
)

type User struct {
	ID      int64
	Name    string
	Age     int
	Address UserAddressResolver
}

type UserAddress struct {
	UserID   int64
	PostCode string
	City     string
	Address1 string
	Address2 string
}

type UserAddressResolver func() (*UserAddress, error)

func (resolver UserAddressResolver) MarshalJSON() ([]byte, error) {
	address, err := resolver()
	if err != nil {
		return nil, err
	}
	return json.Marshal(address)
}

func (resolver UserAddressResolver) ResolveQueryJSON(q *json.Query) (interface{}, error) {
	// validate or rewrite json.Query
	//
	address, err := resolver()
	if err != nil {
		return nil, err
	}
	return address, nil
}

type UserRepository struct{}

func (r UserRepository) FindByID(id int64) (*User, error) {
	v := User{ID: id, Name: "Ken", Age: 20}
	// resolve relation from User to UserAddress
	uaRepo := new(UserAddressRepository)
	v.Address = func() (*UserAddress, error) {
		return uaRepo.FindByUserID(v.ID)
	}
	return v, nil
}

type UserAddressRepository struct{}

func (r UserAddressRepository) FindByUserID(id int64) (*UserAddress, error) {
	return &UserAddress{UserID: id, City: "A", Address1: "hoge", Address2: "fuga"}, nil
}

func main() {
	user, err := new(UserRepository).FindByID(1)
	if err != nil {
		panic(err)
	}
	{
		b, err := json.Marshal(user)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}
	{
		//json.QueryFromJSON(`["Name", "Age", { "Address": [ "City" ] }]`)
		q := json.NewQuery().Fields(
			"Name",
			"Age",
			json.NewQuery("Address").Fields(
				"City",
			),
		)
		b, err := json.MarshalWithQuery(user, q)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}
}
