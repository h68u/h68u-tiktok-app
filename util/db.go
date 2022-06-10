package util

import (
	"tikapp/common/db"
	"tikapp/common/model"
)

type ModelT interface {
	model.User | model.VideoFavorite | model.Video | model.Follow | model.Comment
}

/*func DivideTable[T ModelT](index int64, t T, data interface{}) error {

}*/

func insertIntoTable[T ModelT](t T, data interface{}) error {
	err := db.MySQL.Debug().Model(t).Create(data).Error
	if err != nil {
		return err
	} else {
		return nil
	}
}
