package db

import (
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
)

type UserModel struct {
	CreatedByUID      null.Int    `kuu:"创建人ID"`
	CreatedByUsername null.String `kuu:"创建人账号"`
	CreatedByName     null.String `kuu:"创建人名称"`
	CreatedByAvatar   null.String `kuu:"创建人头像"`
	CreatedByType     null.String `kuu:"创建人类型"`

	UpdatedByUID      null.Int    `kuu:"最后更新人ID"`
	UpdatedByUsername null.String `kuu:"最后更新人账号"`
	UpdatedByName     null.String `kuu:"最后更新人名称"`
	UpdatedByAvatar   null.String `kuu:"最后更新人头像"`
	UpdatedByType     null.String `kuu:"最后更新人类型"`

	DeletedByUID      null.Int    `kuu:"删除人ID"`
	DeletedByUsername null.String `kuu:"删除人账号"`
	DeletedByName     null.String `kuu:"删除人名称"`
	DeletedByAvatar   null.String `kuu:"删除人头像"`
	DeletedByType     null.String `kuu:"删除人类型"`
}

type Model struct {
	ID        uint           `kuu:"主键" gorm:"primarykey"`
	CreatedAt time.Time      `kuu:"创建时间"`
	UpdatedAt time.Time      `kuu:"最后更新时间"`
	DeletedAt gorm.DeletedAt `kuu:"删除时间" gorm:"index"`

	CreatedByUID      null.Int    `kuu:"创建人ID"`
	CreatedByUsername null.String `kuu:"创建人账号"`
	CreatedByName     null.String `kuu:"创建人名称"`
	CreatedByAvatar   null.String `kuu:"创建人头像"`
	CreatedByType     null.String `kuu:"创建人类型"`

	UpdatedByUID      null.Int    `kuu:"最后更新人ID"`
	UpdatedByUsername null.String `kuu:"最后更新人账号"`
	UpdatedByName     null.String `kuu:"最后更新人名称"`
	UpdatedByAvatar   null.String `kuu:"最后更新人头像"`
	UpdatedByType     null.String `kuu:"最后更新人类型"`

	DeletedByUID      null.Int    `kuu:"删除人ID"`
	DeletedByUsername null.String `kuu:"删除人账号"`
	DeletedByName     null.String `kuu:"删除人名称"`
	DeletedByAvatar   null.String `kuu:"删除人头像"`
	DeletedByType     null.String `kuu:"删除人类型"`

	Remark null.String `kuu:"备注"`
}
