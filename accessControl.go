package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"os"
)

type Permission uint8

func (p Permission) MarshalJSON() ([]byte, error) {
	var permissions []string
	if p == Admin {
		permissions = append(permissions, "admin")
		return json.Marshal(permissions)
	}
	if p&Read != 0 {
		permissions = append(permissions, "read")
	}
	if p&Write != 0 {
		permissions = append(permissions, "write")
	}
	if p&CreateTopic != 0 {
		permissions = append(permissions, "createTopic")
	}
	if p&DeleteTopic != 0 {
		permissions = append(permissions, "deleteTopic")
	}
	if p&CreateNewToken != 0 {
		permissions = append(permissions, "createToken")
	}
	if p&DeleteToken != 0 {
		permissions = append(permissions, "deleteToken")
	}
	if p&Subs != 0 {
		permissions = append(permissions, "subscribe")
	}
	return json.Marshal(permissions)
}

var tokens Tokens

type Tokens map[[32]byte]*Token

func (t *Tokens) Save() {
	file, err := os.OpenFile("tokens", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = gob.NewEncoder(file).Encode(t)
	if err != nil {
		panic(err)
	}
}

func (t *Tokens) GetToken(tokenString string) (*Token, bool) {
	tokenHash := sha256.Sum256([]byte(tokenString))
	if token, ok := (*t)[tokenHash]; ok {
		return token, true
	} else {
		return &Token{}, false
	}
}

func (t *Tokens) AddToken(token *Token) {
	tokenHash := sha256.Sum256([]byte(token.Token))
	(*t)[tokenHash] = token
	t.Save()
}

func (t *Tokens) RemoveToken(tokenString string) {
	tokenHash := sha256.Sum256([]byte(tokenString))
	delete(*t, tokenHash)
	t.Save()
}

func LoadTokens() {
	file, err := os.OpenFile("tokens", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = gob.NewDecoder(file).Decode(&tokens)
	if err != nil {
		panic(err)
	}
}

const (
	Read Permission = 1 << iota
	Write
	Subs
	CreateTopic
	DeleteTopic
	CreateNewToken
	DeleteToken
	AdminFlag
	Admin = Read | Write | Subs | CreateTopic | DeleteTopic | CreateNewToken | DeleteToken | AdminFlag
)

type Token struct {
	Token      string     `json:"token_key"`
	Permission Permission `json:"permission"`
	Topics     []string   `json:"subscriptions"`
}

func NewToken(permission Permission, topics ...string) Token {
	newToken := Token{
		Token:      rand.Text(),
		Permission: permission,
		Topics:     topics,
	}
	tokens.AddToken(&newToken)
	return newToken
}

func (token *Token) Subscribe(topic string) {
	if !token.IsSubscribed(topic) {
		token.Topics = append(token.Topics, topic)
	}
	tokens.Save()
}

func (token *Token) HasPermission(permission Permission) bool {
	return token.Permission&permission != 0
}

func (token *Token) IsSubscribed(t string) bool {
	if token.HasPermission(AdminFlag) {
		return true
	}
	for _, topic := range token.Topics {
		if topic == t {
			return true
		}
	}
	return false
}

func InitTokens() Token {
	tokens = make(Tokens)
	admin := NewToken(Admin)
	return admin
}
