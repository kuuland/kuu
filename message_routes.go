package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Messages []Message

func (m Messages) Len() int {
	return len(m)
}

func (m Messages) Less(i, j int) bool {
	return m[i].CreatedAt.Before(m[j].CreatedAt)
}

func (m Messages) Swap(i, j int) {
	tmp := m[i]
	m[i] = m[j]
	m[j] = tmp
}

var MessagesLatestRoute = RouteInfo{
	Name:   "查询当前用户最新消息",
	Method: http.MethodGet,
	Path:   "/messages/latest",
	IntlMessages: map[string]string{
		"messages_latest_failed": "Get latest messages failed.",
	},
	HandlerFunc: func(c *Context) *STDReply {
		c.IgnoreAuth()
		defer c.IgnoreAuth(true)

		var query struct {
			Limit        string `form:"limit"`
			RecipientIDs string `form:"recipient_ids"`
		}
		if err := c.ShouldBindQuery(&query); err != nil {
			return c.STDErr(err, "messages_latest_failed")
		}
		var (
			limit        int
			recipientIDs []uint
		)
		if s := c.DefaultQuery("limit", "10"); s != "" {
			if v, err := strconv.Atoi(s); err == nil {
				limit = v
			}
		}
		if s := c.Query("recipient_ids"); s != "" {
			ss := strings.Split(s, ",")
			for _, item := range ss {
				item = strings.TrimSpace(item)
				if item == "" {
					continue
				}
				if v, err := strconv.Atoi(item); err == nil {
					recipientIDs = append(recipientIDs, uint(v))
				}
			}
		}
		type replyItem struct {
			Messages    Messages
			UnreadCount int
		}
		var reply struct {
			replyItem
			RecipientMap map[uint]replyItem `json:",omitempty"`
		}
		if len(recipientIDs) > 0 {
			reply.RecipientMap = make(map[uint]replyItem)
			for _, itemId := range recipientIDs {
				baseMessageDB := c.DB().Model(&Message{}).Where("sender_id = ? OR recipient_user_ids LIKE ?", itemId, "%"+fmt.Sprintf("%d", itemId)+"%")
				messsagesDB := GetMessageCommonDB(baseMessageDB, c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode, 1, limit)
				if _, v, err := c.ParseCond(&Message{}, messsagesDB); err != nil {
					return c.STDErr(err, "messages_latest_failed")
				} else {
					messsagesDB = v
				}
				var messages Messages
				if err := messsagesDB.Preload("Attachments").Find(&messages).Error; err != nil {
					return c.STDErr(err, "messages_latest_failed")
				}
				sort.Sort(messages)
				item := replyItem{Messages: messages}
				count, err := FindUnreadMessagesCount(baseMessageDB, c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode)
				if err != nil {
					return c.STDErr(err, "messages_latest_failed")
				}
				item.UnreadCount = count
				reply.RecipientMap[itemId] = item
			}
		} else {
			messsagesDB := GetMessageCommonDB(c.DB().Model(&Message{}), c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode, 1, limit)
			if _, v, err := c.ParseCond(&Message{}, messsagesDB); err != nil {
				return c.STDErr(err, "messages_latest_failed")
			} else {
				messsagesDB = v
			}
			var messages Messages
			if err := messsagesDB.Preload("Attachments").Find(&messages).Error; err != nil {
				return c.STDErr(err, "messages_latest_failed")
			}
			sort.Sort(messages)
			reply.Messages = messages
		}
		count, err := FindUnreadMessagesCount(c.DB().Model(&Message{}), c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode)
		if err != nil {
			return c.STDErr(err, "messages_latest_failed")
		}
		reply.UnreadCount = count
		return c.STD(&reply)
	},
}

var MessagesReadRoute = RouteInfo{
	Name:   "阅读消息",
	Method: http.MethodPost,
	Path:   "/messages/read",
	IntlMessages: map[string]string{
		"messages_read_failed": "Update message status failed.",
	},
	HandlerFunc: func(c *Context) *STDReply {
		c.IgnoreAuth()
		defer c.IgnoreAuth(true)

		var body struct {
			MessageIDs   []uint
			RecipientIDs []uint
			All          bool
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			return c.STDErr(err, "messages_read_failed")
		}
		if !body.All && len(body.MessageIDs) == 0 && len(body.RecipientIDs) == 0 {
			return c.STDOK()
		}
		err := c.WithTransaction(func(tx *gorm.DB) error {
			messageDB := tx.Model(&Message{})
			if !body.All {
				if len(body.MessageIDs) > 0 {
					messageDB = messageDB.Where(fmt.Sprintf("%s IN (?)", messageDB.Dialect().Quote("id")), body.MessageIDs)
				}
				if len(body.RecipientIDs) > 0 {
					var (
						sqls  []string
						attrs []interface{}
					)
					for _, itemId := range body.RecipientIDs {
						sqls = append(sqls, "(sender_id = ? OR recipient_user_ids LIKE ?)")
						attrs = append(attrs, itemId, "%"+fmt.Sprintf("%d", itemId)+"%")
					}
					messageDB = messageDB.Where(strings.Join(sqls, " OR "), attrs...)
				}
			}
			messageIDs, err := FindReadableMessageIDs(messageDB, c.SignInfo.UID, c.PrisDesc.ReadableOrgIDs, c.PrisDesc.RolesCode, 0, 0)
			if err != nil {
				return err
			}
			for _, item := range messageIDs {
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

func GetMessageCommonDB(messageDB *gorm.DB, uid uint, orgIDs []uint, rolesCode []string, page, size int) *gorm.DB {
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

	db := messageDB.Where(strings.Join(sqls, " OR "), attrs...)
	db = db.Order(fmt.Sprintf("%s DESC", db.Dialect().Quote("created_at")))
	if size > 0 {
		db = db.Limit(size)
	}
	if page > 0 && size > 0 {
		db = db.Offset((page - 1) * size)
	}
	return db
}

func FindReadableMessageIDs(messageDB *gorm.DB, uid uint, orgIDs []uint, rolesCode []string, page, size int) ([]uint, error) {
	db := GetMessageCommonDB(messageDB, uid, orgIDs, rolesCode, page, size)

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

func FindUnreadMessages(messageDB *gorm.DB, uid uint, orgIDs []uint, rolesCode []string, page, size int) ([]Message, error) {
	db, err := FindUnreadMessagesPrepareDB(messageDB, uid, orgIDs, rolesCode)
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

func FindUnreadMessagesCount(messageDB *gorm.DB, uid uint, orgIDs []uint, rolesCode []string) (int, error) {
	db, err := FindUnreadMessagesPrepareDB(messageDB, uid, orgIDs, rolesCode)
	if err != nil {
		return 0, err
	}
	var count int
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func FindUnreadMessagesPrepareDB(messageDB *gorm.DB, uid uint, orgIDs []uint, rolesCode []string) (*gorm.DB, error) {
	messageIDs, err := FindReadableMessageIDs(messageDB, uid, orgIDs, rolesCode, 0, 0)
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
	messageDB = messageDB.Where("sender_id <> ?", uid).Where(fmt.Sprintf("%s IN (?)", DB().Dialect().Quote("id")), messageIDs)
	if len(readMessageIDs) > 0 {
		messageDB = messageDB.Where(fmt.Sprintf("%s NOT IN (?)", DB().Dialect().Quote("id")), readMessageIDs)
	}
	return messageDB, nil
}
