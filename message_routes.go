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
		messages, err := findReadableMessages(c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode)
		if err != nil {
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
		messages, err := findUnreadMessages(c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode)
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
		var body struct {
			MessageIDs []uint `binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			return c.STDErr(err, "messages_read_failed")
		}
		err := c.WithTransaction(func(tx *gorm.DB) error {
			readableMessageIDs, err := findReadableMessageIDs(c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode)
			if err != nil {
				return err
			}
			readableMessageIDMap := make(map[uint]bool)
			for _, item := range readableMessageIDs {
				readableMessageIDMap[item] = true
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

func findReadableMessageIDs(uid uint, orgIDs []uint, rolesCode []string) ([]uint, error) {
	var (
		sqls  []string
		attrs []interface{}
	)
	sqls = append(sqls, fmt.Sprintf("%s LIKE ?", DB().Dialect().Quote("user_ids")))
	attrs = append(attrs, "%,"+strconv.Itoa(int(uid))+",%")
	for _, orgId := range orgIDs {
		sqls = append(sqls, fmt.Sprintf("%s LIKE ?", DB().Dialect().Quote("org_ids")))
		attrs = append(attrs, "%,"+strconv.Itoa(int(orgId))+",%")
	}
	for _, roleCode := range rolesCode {
		sqls = append(sqls, fmt.Sprintf("%s LIKE ?", DB().Dialect().Quote("role_codes")))
		attrs = append(attrs, "%,"+roleCode+",%")
	}

	var ranges []MessageRange
	if err := DB().Model(&MessageRange{}).Select(fmt.Sprintf("%s", DB().Dialect().Quote("message_id"))).Where(strings.Join(sqls, "OR"), attrs...).Find(&ranges).Error; err != nil {
		return nil, err
	}
	var (
		messageIDCache = make(map[uint]bool)
		messageIDs     []uint
	)
	for _, item := range ranges {
		if messageIDCache[item.MessageID] {
			continue
		}
		messageIDCache[item.MessageID] = true
		messageIDs = append(messageIDs, item.MessageID)
	}
	return messageIDs, nil
}

func findReadableMessages(uid uint, orgIDs []uint, rolesCode []string) ([]Message, error) {
	messageIDs, err := findReadableMessageIDs(uid, orgIDs, rolesCode)
	if err != nil {
		return nil, err
	}
	var messages []Message
	if err := DB().Model(&Message{}).Where(fmt.Sprintf("%s IN (?)", DB().Dialect().Quote("id")), messageIDs).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func findUnreadMessages(uid uint, orgIDs []uint, rolesCode []string) ([]Message, error) {
	db, err := findUnreadMessagesPrepareDB(uid, orgIDs, rolesCode)
	if err != nil {
		return nil, err
	}
	var messages []Message
	if err := db.Find(&messages).Error; err != nil {
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
	messageIDs, err := findReadableMessageIDs(uid, orgIDs, rolesCode)
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
