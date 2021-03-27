package main

import (
	"context"
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

type UserRepository struct {
	uaRepo *UserAddressRepository
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		uaRepo: NewUserAddressRepository(),
	}
}

type UserAddressRepository struct{}

func NewUserAddressRepository() *UserAddressRepository {
	return &UserAddressRepository{}
}

type UserAddressResolver func(context.Context) (*UserAddress, error)

func (resolver UserAddressResolver) MarshalJSON(ctx context.Context) ([]byte, error) {
	address, err := resolver(ctx)
	if err != nil {
		return nil, err
	}
	return json.MarshalWithQuery(address, json.QueryFromContext(ctx))
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*User, error) {
	v := User{ID: id, Name: "Ken", Age: 20}
	// resolve relation from User to UserAddress
	v.Address = func(ctx context.Context) (*UserAddress, error) {
		return r.uaRepo.FindByUserID(ctx, v.ID)
	}
	return v, nil
}

func (*UserAddressRepository) FindByUserID(ctx context.Context, id int64) (*UserAddress, error) {
	return &UserAddress{
		UserID:   id,
		City:     "A",
		Address1: "hoge",
		Address2: "fuga",
	}, nil
}

func main() {
	userRepo := NewUserRepository()
	user, err := userRepo.FindByID(context.Background(), 1)
	if err != nil {
		panic(err)
	}
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
