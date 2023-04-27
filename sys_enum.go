package kuu

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	"github.com/samber/lo"
	"gopkg.in/guregu/null.v3"
)

var (
	enumMap            = make(map[string]*Enum)
	enumSyncChannelKey = "kuu_enum_sync_channel_key" // 主应用同步给集群
	enumKey            = "kuu_enums"
	enumLocalKeys      []string
)

func init() {
	if !C().Has("configRedisServer") {
		return
	}
	// 获取集群的enums
	m := DefaultConfigServer.HGetAll(context.Background(), enumKey).Val()
	for key, value := range m {
		var desc Enum
		if err := JSONParse(value, &desc); err != nil || desc.ID == 0 {
			continue
		}
		desc.gen()
		enumMap[key] = &desc
	}
	// 监听集群的注册
	ctx := context.Background()
	// 监听集群的变化
	ps := DefaultConfigServer.Subscribe(ctx, enumSyncChannelKey)
	if _, err := ps.Receive(ctx); err != nil {
		return
	}
	ch := ps.Channel()
	go func(ch <-chan *redis.Message) {
		for msg := range ch {
			value := msg.Payload
			var desc Enum
			if err := JSONParse(value, &desc); err != nil {
				ERROR("Enum sync Error: %s", err.Error())
				continue
			}
			if !lo.Contains(enumLocalKeys, desc.Code) {
				continue
			}
			INFO("Enum loaded from config server: %s(%s)", desc.Code, desc.Name)
			desc.gen()
			enumMap[desc.Code] = &desc
		}
	}(ch)
}

type Enum struct {
	gorm.Model
	Code   string
	Name   string
	Values map[string]string `json:"-" gorm:"-"`
	Alias  map[string]string `json:"-" gorm:"-"`
	Items  []EnumItem        `json:"Values"`
}

type EnumItem struct {
	gorm.Model
	EnumID   uint
	Label    string
	Alias    string
	Value    string
	Disabled null.Bool
}

func (enum *Enum) gen() {
	enum.Alias = make(map[string]string, len(enum.Items))
	enum.Values = make(map[string]string, len(enum.Items))
	for _, item := range enum.Items {
		enum.Alias[item.Alias] = item.Value
		enum.Values[item.Value] = item.Label
	}
}

func (enum *Enum) AfterSave(tx *gorm.DB) error {
	go func() {
		if C().Has(DefaultConfigServerKey) {
			ctx := context.Background()
			DefaultConfigServer.HSet(ctx, enumKey, enum.Code, JSONStringify(enum))
			DefaultConfigServer.Publish(ctx, enumSyncChannelKey, JSONStringify(enum))
		}
	}()
	return nil
}

func EnumMap() map[string]*Enum {
	return enumMap
}

func EnumList() (list []*Enum) {
	for _, item := range EnumMap() {
		list = append(list, item)
	}
	return
}

func GetEnumItem(classCode string) *Enum {
	return enumMap[classCode]
}

func GetEnumLabel(classCode string, value string) (label string) {
	item := GetEnumItem(classCode)
	if item != nil {
		label = item.Values[value]
	}
	return
}

func GetEnumValue(classCode string, alias string) string {
	item := GetEnumItem(classCode)
	if item != nil {
		return item.Alias[alias]
	}
	return ""
}

func loadEnumToConfigServer() {
	if !checkConfigServer() {
		return
	}
	var list []Enum
	DB().Model(&Enum{}).Preload("Items").Find(&list)
	ctx := context.Background()
	DefaultConfigServer.Del(ctx, enumKey)
	for _, desc := range list {
		DefaultConfigServer.HSet(ctx, enumKey, desc.Code, JSONStringify(desc))
		DefaultConfigServer.Publish(ctx, enumSyncChannelKey, JSONStringify(desc))
	}
}
