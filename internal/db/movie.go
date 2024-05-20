package db

import (
	"github.com/synctv-org/synctv/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateMovie(movie *model.Movie) error {
	return db.Create(movie).Error
}

func CreateMovies(movies []*model.Movie) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(movies).Error
	})
}

func WithParentMovieID(parentMovieID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if parentMovieID == "" {
			return db.Where("base_parent_id IS NULL")
		}
		return db.Where("base_parent_id = ?", parentMovieID)
	}
}

func GetMoviesByRoomID(roomID string, scopes ...func(*gorm.DB) *gorm.DB) ([]*model.Movie, error) {
	movies := []*model.Movie{}
	err := db.Where("room_id = ?", roomID).Order("position ASC").Scopes(scopes...).Find(&movies).Error
	return movies, err
}

func GetMoviesCountByRoomID(roomID string, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&model.Movie{}).Where("room_id = ?", roomID).Scopes(scopes...).Count(&count).Error
	return count, err
}

func GetMovieByID(roomID, id string, scopes ...func(*gorm.DB) *gorm.DB) (*model.Movie, error) {
	movie := &model.Movie{}
	err := db.Where("room_id = ? AND id = ?", roomID, id).Scopes(scopes...).First(movie).Error
	return movie, HandleNotFound(err, "room or movie")
}

func DeleteMovieByID(roomID, id string) error {
	err := db.Unscoped().Where("room_id = ? AND id = ?", roomID, id).Delete(&model.Movie{}).Error
	return HandleNotFound(err, "room or movie")
}

func DeleteMoviesByID(roomID string, ids []string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().Where("room_id = ? AND id IN ?", roomID, ids).Delete(&model.Movie{}).Error
		if err != nil {
			return HandleNotFound(err, "room or movie")
		}
		return nil
	})
}

func LoadAndDeleteMovieByID(roomID, id string, columns []clause.Column) (*model.Movie, error) {
	movie := &model.Movie{}
	err := db.Unscoped().Clauses(clause.Returning{Columns: columns}).Where("room_id = ? AND id = ?", roomID, id).Delete(movie).Error
	return movie, HandleNotFound(err, "room or movie")
}

func DeleteMoviesByRoomID(roomID string, scopes ...func(*gorm.DB) *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("room_id = ?", roomID).Scopes(scopes...).Delete(&model.Movie{}).Error
		if err != nil {
			return HandleNotFound(err, "room")
		}
		return nil
	})
}

func DeleteMoviesByRoomIDAndParentID(roomID string, parentID string) error {
	return DeleteMoviesByRoomID(roomID, WithParentMovieID(parentID))
}

func LoadAndDeleteMoviesByRoomID(roomID string, columns ...clause.Column) ([]*model.Movie, error) {
	movies := []*model.Movie{}
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().Clauses(clause.Returning{Columns: columns}).Where("room_id = ?", roomID).Delete(&movies).Error
		return HandleNotFound(err, "room")
	})
	return movies, err
}

func UpdateMovie(movie *model.Movie, columns ...clause.Column) error {
	err := db.Model(movie).Clauses(clause.Returning{Columns: columns}).Where("room_id = ? AND id = ?", movie.RoomID, movie.ID).Updates(movie).Error
	return HandleNotFound(err, "room or movie")
}

func SaveMovie(movie *model.Movie, columns ...clause.Column) error {
	err := db.Model(movie).Clauses(clause.Returning{Columns: columns}).Where("room_id = ? AND id = ?", movie.RoomID, movie.ID).Omit("created_at").Save(movie).Error
	return HandleNotFound(err, "room or movie")
}

func SwapMoviePositions(roomID, movie1ID, movie2ID string) (err error) {
	return Transactional(func(tx *gorm.DB) error {
		movie1 := &model.Movie{}
		movie2 := &model.Movie{}
		err = tx.Where("room_id = ? AND id = ?", roomID, movie1ID).First(movie1).Error
		if err != nil {
			return HandleNotFound(err, "movie1")
		}
		err = tx.Where("room_id = ? AND id = ?", roomID, movie2ID).First(movie2).Error
		if err != nil {
			return HandleNotFound(err, "movie2")
		}
		movie1.Position, movie2.Position = movie2.Position, movie1.Position
		err = tx.Omit("created_at").Save(movie1).Error
		if err != nil {
			return err
		}
		return tx.Omit("created_at").Save(movie2).Error
	})
}
