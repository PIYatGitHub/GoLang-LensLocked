package models

import (
	"errors"

	"github.com/jinzhu/gorm"

	"../hash"
	"../rand"
	// added it not to get confused as of what is needed to run this...
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

var (
	//ErrNotFound is returned whenever you cannot find the resource at the DB
	ErrNotFound = errors.New("models: resource not found")
	// ErrInvalidID is returned if you attempt to pass in an Id <= 0
	ErrInvalidID = errors.New("models: the ID is supposed to be greater than 0")
	// ErrInvalidPass is returned if you passed in a wrong password
	ErrInvalidPass = errors.New("models: the password provided is invalid")
)

const userPwP = "wrjg82j8#$%^&#Rweg4128y8y8suTO(24#%9ghsdbu"
const hmacSecretKey = "4wjht8wywr!^Y@$Yggwj8qeyrh139hSFYHEYFehjeo235"

// NewUserService will open a singular connection to the DB!
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	hmac := hash.NewHMAC(hmacSecretKey)
	return &UserService{
		db:   db,
		hmac: hmac,
	}, nil
}

// UserService holds the logic?
type UserService struct {
	db   *gorm.DB
	hmac hash.HMAC
}

//ByID will lookup the user by id;
// it will return user,nil or nil for the user and specific user (only one)
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

//ByEmail will lookup the user by his/her email address;
// it will return user,nil or nil for the user and specific user (only one)
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

//ByRemember will lookup the user by his/her remember token;
// it will return user,nil or nil for the user and specific user (only one)
//the method will handle the hashing for us as well
func (us *UserService) ByRemember(token string) (*User, error) {
	var user User
	rememberHash := us.hmac.Hash(token)
	err := first(us.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

//Authenticate will lookup the provided email and pass and will return
//a user obj for logged user and err if there isnt a user
func (us *UserService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+userPwP))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrInvalidPass
		default:
			return nil, err
		}
	}
	return foundUser, err
}

// first is a function to get the first match from the DB.
// DO NOT FORGET to give it a pointer on the dst object, otherwise
//you may run into major pizdec!
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

//Create does take care of creating a user or returns an error if there is sth wrong...
func (us *UserService) Create(user *User) error {
	pwBytes := []byte(user.Password + userPwP)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	if user.Remember == "" {
		toekn, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = toekn
	}
	user.RememberHash = us.hmac.Hash(user.Remember)

	return us.db.Create(user).Error
}

//Update does take care of updating a user or returns an error if there is sth wrong...
func (us *UserService) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = us.hmac.Hash(user.Remember)
	}
	return us.db.Save(user).Error
}

//Delete is a dangerous function as it deletes the user by ID. Do not use it if you are not sure...
func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

// Close will terminate the connection to the DB!
func (us *UserService) Close() error {
	return us.db.Close()
}

//DestructiveReset deletes the users table. NEVER EVER RUN IN PROD!!!!!
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}

//AutoMigrate is our version of the GORM function. We will use it further down the line
func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

// User will serve to save our users with the appropriate fields...
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null; unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null; unique_index"`
}
