package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var MessagesRoute = RouteInfo{
	Name:   "查询当前用户所有消息",
	Method: http.MethodGet,
	Path:   "/messages",
	IntlMessages: map[string]string{
		"messages_failed": "Get messages failed.",
	},
	HandlerFunc: func(c *Context) *STDReply {
		c.IgnoreAuth()
		defer c.IgnoreAuth(true)

		page, size := c.GetPagination(true)
		db := getMessageCommonDB(c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode, page, size)
		if _, v, err := c.ParseCond(&Message{}, db); err != nil {
			return c.STDErr(err, "messages_failed")
		} else {
			db = v
		}
		var messages []Message
		if err := db.Preload("Attachments").Find(&messages).Error; err != nil {
			return c.STDErr(err, "messages_failed")
		}
		return c.STD(messages)
	},
}

var MessagesUnreadRoute = RouteInfo{
	Name:   "查询当前用户所有未读消息",
	Method: http.MethodGet,
	Path:   "/messages/unread",
	IntlMessages: map[string]string{
		"messages_unread_failed": "Get unread messages failed.",
	},
	HandlerFunc: func(c *Context) *STDReply {
		c.IgnoreAuth()
		defer c.IgnoreAuth(true)

		page, size := c.GetPagination()
		messages, err := findUnreadMessages(c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode, page, size)
		if err != nil {
			return c.STDErr(err, "messages_unread_count_failed")
		}
		return c.STD(messages)
	},
}

var MessagesUnreadCountRoute = RouteInfo{
	Name:   "查询当前用户所有未读消息总数",
	Method: http.MethodGet,
	Path:   "/messages/unread/count",
	IntlMessages: map[string]string{
		"messages_unread_count_failed": "Get unread messages count failed.",
	},
	HandlerFunc: func(c *Context) *STDReply {
		c.IgnoreAuth()
		defer c.IgnoreAuth(true)

		count, err := findUnreadMessagesCount(c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode)
		if err != nil {
			return c.STDErr(err, "messages_unread_count_failed")
		}
		return c.STD(count)
	},
}

var MessagesReadRoute = RouteInfo{
	Name:   "阅读消息",
	Method: http.MethodPost,
	Path:   "/messages/read",
	IntlMessages: map[string]string{
		"messages_read_failed": "Read status update failed.",
	},
	HandlerFunc: func(c *Context) *STDReply {
		c.IgnoreAuth()
		defer c.IgnoreAuth(true)

		var body struct {
			MessageIDs []uint `binding:"required"`
			All        bool
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			return c.STDErr(err, "messages_read_failed")
		}
		err := c.WithTransaction(func(tx *gorm.DB) error {
			readableMessageIDs, err := findReadableMessageIDs(c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode, 0, 0)
			if err != nil {
				return err
			}
			readableMessageIDMap := make(map[uint]bool)
			for _, item := range readableMessageIDs {
				readableMessageIDMap[item] = true
			}
			if body.All {
				body.MessageIDs = readableMessageIDs
			}
			for _, item := range body.MessageIDs {
				if !readableMessageIDMap[item] {
					continue
				}
				if err := tx.Model(&MessageReceipt{}).Create(&MessageReceipt{
					MessageID:         item,
					RecipientID:       c.SignInfo.UID,
					RecipientUsername: c.SignInfo.Username,
					RecipientSourceIP: c.ClientIP(),
					ReadAt:            null.TimeFrom(time.Now()),
				}).Error; err != nil {
					return err
				}
			}
			return tx.Error
		})
		if err != nil {
			return c.STDErr(err, "messages_read_failed")
		}
		return c.STDOK()
	},
}

func getMessageCommonDB(uid uint, orgIDs []uint, rolesCode []string, page, size int) *gorm.DB {
	var (
		sqls  []string
		attrs []interface{}
	)
	// 指定用户的消息
	sqls = append(sqls, fmt.Sprintf("%s LIKE ?", DB().Dialect().Quote("recipient_user_ids")))
	attrs = append(attrs, "%,"+strconv.Itoa(int(uid))+",%")
	// 指定组织的消息
	for _, orgId := range orgIDs {
		sqls = append(sqls, fmt.Sprintf("%s LIKE ?", DB().Dialect().Quote("recipient_org_ids")))
		attrs = append(attrs, "%,"+strconv.Itoa(int(orgId))+",%")
	}
	// 指定角色的消息
	for _, roleCode := range rolesCode {
		sqls = append(sqls, fmt.Sprintf("%s LIKE ?", DB().Dialect().Quote("recipient_role_codes")))
		attrs = append(attrs, "%,"+roleCode+",%")
	}
	// 指定创建人/发送人的消息
	sqls = append(sqls, fmt.Sprintf("%s = ?", DB().Dialect().Quote("created_by_id")))
	attrs = append(attrs, uid)
	sqls = append(sqls, fmt.Sprintf("%s = ?", DB().Dialect().Quote("sender_id")))
	attrs = append(attrs, uid)

	db := DB().Model(&Message{}).Where(strings.Join(sqls, " OR "), attrs...)
	db = db.Order(fmt.Sprintf("%s DESC", db.Dialect().Quote("created_at")))
	if size > 0 {
		db = db.Limit(size)
	}
	if page > 0 && size > 0 {
		db = db.Offset((page - 1) * size)
	}
	return db
}

func findReadableMessageIDs(uid uint, orgIDs []uint, rolesCode []string, page, size int) ([]uint, error) {
	db := getMessageCommonDB(uid, orgIDs, rolesCode, page, size)

	var messages []Message
	if err := db.Select(fmt.Sprintf("%s", DB().Dialect().Quote("id"))).Find(&messages).Error; err != nil {
		return nil, err
	}
	var messageIDs []uint
	for _, item := range messages {
		messageIDs = append(messageIDs, item.ID)
	}
	return messageIDs, nil
}

func findReadableMessages(uid uint, orgIDs []uint, rolesCode []string, page, size int) ([]Message, error) {
	db := getMessageCommonDB(uid, orgIDs, rolesCode, page, size)
	var messages []Message
	if err := db.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func findUnreadMessages(uid uint, orgIDs []uint, rolesCode []string, page, size int) ([]Message, error) {
	db, err := findUnreadMessagesPrepareDB(uid, orgIDs, rolesCode)
	if err != nil {
		return nil, err
	}
	if size > 0 {
		db = db.Limit(size)
	}
	if page > 0 && size > 0 {
		db = db.Offset((page - 1) * size)
	}
	var messages []Message
	if err := db.Order(fmt.Sprintf("%s DESC", db.Dialect().Quote("created_at"))).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func findUnreadMessagesCount(uid uint, orgIDs []uint, rolesCode []string) (int, error) {
	db, err := findUnreadMessagesPrepareDB(uid, orgIDs, rolesCode)
	if err != nil {
		return 0, err
	}
	var count int
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func findUnreadMessagesPrepareDB(uid uint, orgIDs []uint, rolesCode []string) (*gorm.DB, error) {
	messageIDs, err := findReadableMessageIDs(uid, orgIDs, rolesCode, 0, 0)
	if err != nil {
		return nil, err
	}
	var receipts []MessageReceipt
	if err := DB().Model(&MessageReceipt{}).
		Select(fmt.Sprintf("%s", DB().Dialect().Quote("message_id"))).
		Where(fmt.Sprintf("%s = ?", DB().Dialect().Quote("recipient_id")), uid).
		Find(&receipts).Error; err != nil {
		return nil, err
	}
	var readMessageIDs []uint
	for _, item := range receipts {
		readMessageIDs = append(readMessageIDs, item.MessageID)
	}
	db := DB().Model(&Message{}).
		Where(fmt.Sprintf("%s IN (?)", DB().Dialect().Quote("id")), messageIDs).
		Where(fmt.Sprintf("%s NOT IN (?)", DB().Dialect().Quote("id")), readMessageIDs)
	return db, nil
}
