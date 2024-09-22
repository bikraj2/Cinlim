package data

import (
	"database/sql"
	"errors"
)

var (
	ErrNoRecordFound = errors.New("record not Found")
	ErrEditConflict  = errors.New("eidt conflict")
)

type Models struct {
	Movies     MovieModel
	Users      UserModel
	Tokens     TokenModel
	Permission PermissionModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies:     MovieModel{DB: db},
		Permission: PermissionModel{DB: db},
		Tokens:     TokenModel{DB: db},
		Users:      UserModel{DB: db},
	}
}
