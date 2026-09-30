package json_test

import (
	"context"
	"fmt"
	"log"

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
	return json.MarshalContext(ctx, address)
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*User, error) {
	user := &User{ID: id, Name: "Ken", Age: 20}
	// resolve relation from User to UserAddress
	user.Address = func(ctx context.Context) (*UserAddress, error) {
		return r.uaRepo.FindByUserID(ctx, user.ID)
	}
	return user, nil
}

func (*UserAddressRepository) FindByUserID(ctx context.Context, id int64) (*UserAddress, error) {
	return &UserAddress{
		UserID:   id,
		City:     "A",
		Address1: "foo",
		Address2: "bar",
	}, nil
}

func Example_selection() {
	ctx := context.Background()
	userRepo := NewUserRepository()
	user, err := userRepo.FindByID(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}
	sel, err := json.ParseSelection(`Name Age Address { City }`)
	if err != nil {
		log.Fatal(err)
	}
	// Address is resolved by MarshalJSON(context.Context), which is given the selection of the address.
	b, err := json.MarshalContext(ctx, user, json.WithSelection(sel))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	// Output:
	// {"Name":"Ken","Age":20,"Address":{"City":"A"}}
}
