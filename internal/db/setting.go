package db

import (
	"github.com/synctv-org/synctv/internal/model"
	"gorm.io/gorm/clause"
)

func GetSettingItems() ([]*model.Setting, error) {
	var items []*model.Setting
	err := db.Find(&items).Error
	return items, err
}

func GetSettingItemsToMap() (map[string]*model.Setting, error) {
	items, err := GetSettingItems()
	if err != nil {
		return nil, err
	}
	m := make(map[string]*model.Setting)
	for _, item := range items {
		m[item.Name] = item
	}
	return m, nil
}

func GetSettingItemByName(name string) (*model.Setting, error) {
	var item model.Setting
	err := db.Where("name = ?", name).First(&item).Error
	return &item, err
}

func SaveSettingItem(item *model.Setting) error {
	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(item).Error
}

func DeleteSettingItem(item *model.Setting) error {
	return db.Delete(item).Error
}

func DeleteSettingItemByName(name string) error {
	return db.Delete(&model.Setting{Name: name}).Error
}

func GetSettingItemValue(name string) (string, error) {
	var value string
	err := db.Model(&model.Setting{}).Where("name = ?", name).Select("value").First(&value).Error
	return value, err
}

func FirstOrCreateSettingItemValue(s *model.Setting) error {
	return db.Where("name = ?", s.Name).Attrs(model.Setting{
		Value: s.Value,
		Type:  s.Type,
		Group: s.Group,
	}).FirstOrCreate(s).Error
}

func UpdateSettingItemValue(name, value string) error {
	return db.Model(&model.Setting{}).Where("name = ?", name).Update("value", value).Error
}
