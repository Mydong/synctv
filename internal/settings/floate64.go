package settings

import (
	"fmt"
	"math"
	"strconv"
	"sync/atomic"

	"github.com/synctv-org/synctv/internal/db"
	"github.com/synctv-org/synctv/internal/model"
)

type Float64Setting interface {
	Setting
	Set(float64) error
	Get() float64
	Default() float64
	Parse(string) (float64, error)
	Stringify(float64) string
}

var _ Float64Setting = (*Float64)(nil)

type Float64 struct {
	setting
	defaultValue float64
	value        uint64
}

func NewFloat64(name string, value float64, group model.SettingGroup) *Float64 {
	f := &Float64{
		setting: setting{
			name:        name,
			group:       group,
			settingType: model.SettingTypeFloat64,
		},
		defaultValue: value,
	}
	f.set(value)
	return f
}

func (f *Float64) Parse(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func (f *Float64) Stringify(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (f *Float64) Init(value string) error {
	v, err := f.Parse(value)
	if err != nil {
		return err
	}
	f.set(v)
	return nil
}

func (f *Float64) Raw() string {
	return f.Stringify(f.Get())
}

func (f *Float64) Default() float64 {
	return f.defaultValue
}

func (f *Float64) DefaultRaw() string {
	return f.Stringify(f.defaultValue)
}

func (f *Float64) DefaultInterface() any {
	return f.Default()
}

func (f *Float64) SetRaw(value string) error {
	err := f.Init(value)
	if err != nil {
		return err
	}
	return db.UpdateSettingItemValue(f.Name(), f.Raw())
}

func (f *Float64) set(value float64) {
	atomic.StoreUint64(&f.value, math.Float64bits(value))
}

func (f *Float64) Set(value float64) error {
	err := db.UpdateSettingItemValue(f.Name(), f.Stringify(value))
	if err != nil {
		return err
	}
	f.set(value)
	return nil
}

func (f *Float64) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&f.value))
}

func (f *Float64) Interface() any {
	return f.Get()
}

func newFloat64Setting(k string, v float64, g model.SettingGroup) *Float64 {
	_, loaded := Settings[k]
	if loaded {
		panic(fmt.Sprintf("setting %s already exists", k))
	}
	f := NewFloat64(k, v, g)
	Settings[k] = f
	GroupSettings[g] = append(GroupSettings[g], f)
	return f
}
