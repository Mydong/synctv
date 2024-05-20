package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maruel/natural"
	"github.com/sirupsen/logrus"
	"github.com/synctv-org/synctv/internal/db"
	dbModel "github.com/synctv-org/synctv/internal/model"
	"github.com/synctv-org/synctv/internal/op"
	"github.com/synctv-org/synctv/internal/settings"
	"github.com/synctv-org/synctv/server/middlewares"
	"github.com/synctv-org/synctv/server/model"
	"github.com/synctv-org/synctv/utils"
	"github.com/zijiren233/gencontainer/refreshcache"
	"github.com/zijiren233/gencontainer/synccache"
	"gorm.io/gorm"
)

var (
	ErrAuthFailed  = errors.New("auth failed")
	ErrAuthExpired = errors.New("auth expired")
	ErrRoomAlready = errors.New("room already exists")
)

func RoomMe(ctx *gin.Context) {
	user := ctx.MustGet("user").(*op.UserEntry).Value()
	room := ctx.MustGet("room").(*op.RoomEntry).Value()
	log := ctx.MustGet("log").(*logrus.Entry)

	rur, err := room.LoadOrCreateRoomMember(user.ID)
	if err != nil {
		log.Errorf("room me failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(&model.RoomMeResp{
		UserID:           user.ID,
		RoomID:           room.ID,
		JoinAt:           rur.CreatedAt.UnixMilli(),
		Role:             rur.Role,
		Permissions:      rur.Permissions,
		AdminPermissions: rur.AdminPermissions,
	}))
}

func RoomPiblicSettings(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.RoomEntry).Value()
	ctx.JSON(http.StatusOK, model.NewApiDataResp(room.Settings))
}

func CreateRoom(ctx *gin.Context) {
	user := ctx.MustGet("user").(*op.UserEntry).Value()
	log := ctx.MustGet("log").(*logrus.Entry)

	if settings.DisableCreateRoom.Get() && !user.IsAdmin() {
		log.Error("create room is disabled")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorStringResp("create room is disabled"))
		return
	}

	req := model.CreateRoomReq{}
	if err := model.Decode(ctx, &req); err != nil {
		log.Errorf("create room failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	room, err := user.CreateRoom(req.RoomName, req.Password, db.WithSettingHidden(req.Settings.Hidden))
	if err != nil {
		log.Errorf("create room failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	token, err := middlewares.NewAuthRoomToken(user, room.Value())
	if err != nil {
		log.Errorf("create room failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.JSON(http.StatusCreated, model.NewApiDataResp(gin.H{
		"roomId": room.Value().ID,
		"token":  token,
	}))
}

var roomHotCache = refreshcache.NewRefreshCache(func(context.Context, ...any) ([]*model.RoomListResp, error) {
	rooms := make([]*model.RoomListResp, 0)
	op.RangeRoomCache(func(key string, value *synccache.Entry[*op.Room]) bool {
		v := value.Value()
		if !v.Settings.Hidden {
			rooms = append(rooms, &model.RoomListResp{
				RoomId:       v.ID,
				RoomName:     v.Name,
				PeopleNum:    v.PeopleNum(),
				NeedPassword: v.NeedPassword(),
				Creator:      op.GetUserName(v.CreatorID),
				CreatedAt:    v.CreatedAt.UnixMilli(),
			})
		}
		return true
	})

	slices.SortStableFunc(rooms, func(a, b *model.RoomListResp) int {
		if a.PeopleNum == b.PeopleNum {
			if a.RoomName == b.RoomName {
				return 0
			}
			if natural.Less(a.RoomName, b.RoomName) {
				return -1
			} else {
				return 1
			}
		} else if a.PeopleNum > b.PeopleNum {
			return -1
		} else {
			return 1
		}
	})

	return rooms, nil
}, time.Second*3)

func RoomHotList(ctx *gin.Context) {
	log := ctx.MustGet("log").(*logrus.Entry)

	page, pageSize, err := utils.GetPageAndMax(ctx)
	if err != nil {
		log.Errorf("get room hot list failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	r, err := roomHotCache.Get(ctx)
	if err != nil {
		log.Errorf("get room hot list failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"total": len(r),
		"list":  utils.GetPageItems(r, page, pageSize),
	}))
}

func RoomList(ctx *gin.Context) {
	log := ctx.MustGet("log").(*logrus.Entry)

	page, pageSize, err := utils.GetPageAndMax(ctx)
	if err != nil {
		log.Errorf("get room list failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	var desc = ctx.DefaultQuery("order", "desc") == "desc"

	scopes := []func(db *gorm.DB) *gorm.DB{
		func(db *gorm.DB) *gorm.DB {
			return db.InnerJoins("JOIN room_settings ON rooms.id = room_settings.id")
		},
		db.WhereRoomSettingWithoutHidden(),
		db.WhereStatus(dbModel.RoomStatusActive),
	}

	switch ctx.DefaultQuery("sort", "name") {
	case "createdAt":
		if desc {
			scopes = append(scopes, db.OrderByCreatedAtDesc)
		} else {
			scopes = append(scopes, db.OrderByCreatedAtAsc)
		}
	case "name":
		if desc {
			scopes = append(scopes, db.OrderByDesc("name"))
		} else {
			scopes = append(scopes, db.OrderByAsc("name"))
		}
	default:
		log.Errorf("get room list failed: not support sort")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorStringResp("not support sort"))
		return
	}

	if keyword := ctx.Query("keyword"); keyword != "" {
		// search mode, all, name, creator
		switch ctx.DefaultQuery("search", "all") {
		case "all":
			ids, err := db.GerUsersIDByUsernameLike(keyword)
			if err != nil {
				log.Errorf("get room list failed: %v", err)
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
				return
			}
			scopes = append(scopes, db.WhereRoomNameLikeOrCreatorInOrRoomsIDLike(keyword, ids, keyword))
		case "name":
			scopes = append(scopes, db.WhereRoomNameLike(keyword))
		case "creator":
			ids, err := db.GerUsersIDByUsernameLike(keyword)
			if err != nil {
				log.Errorf("get room list failed: %v", err)
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
				return
			}
			scopes = append(scopes, db.WhereCreatorIDIn(ids))
		case "id":
			scopes = append(scopes, db.WhereRoomsIDLike(keyword))
		}
	}

	total, err := db.GetAllRoomsCount(scopes...)
	if err != nil {
		log.Errorf("get room list failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	list, err := genRoomListResp(append(scopes, db.Paginate(page, pageSize))...)
	if err != nil {
		log.Errorf("get room list failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"total": total,
		"list":  list,
	}))
}

func genRoomListResp(scopes ...func(db *gorm.DB) *gorm.DB) ([]*model.RoomListResp, error) {
	rs, err := db.GetAllRooms(scopes...)
	if err != nil {
		return nil, err
	}
	resp := make([]*model.RoomListResp, len(rs))
	for i, r := range rs {
		resp[i] = &model.RoomListResp{
			RoomId:       r.ID,
			RoomName:     r.Name,
			PeopleNum:    op.PeopleNum(r.ID),
			NeedPassword: len(r.HashedPassword) != 0,
			CreatorID:    r.CreatorID,
			Creator:      op.GetUserName(r.CreatorID),
			CreatedAt:    r.CreatedAt.UnixMilli(),
			Status:       r.Status,
		}
	}
	return resp, nil
}

func CheckRoom(ctx *gin.Context) {
	log := ctx.MustGet("log").(*logrus.Entry)

	r, err := db.GetRoomByID(ctx.Query("roomId"))
	if err != nil {
		log.Errorf("check room failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusNotFound, model.NewApiErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"peopleNum":    op.PeopleNum(r.ID),
		"needPassword": r.NeedPassword(),
		"creator":      op.GetUserName(r.CreatorID),
	}))
}

func GuestJoinRoom(ctx *gin.Context) {
	log := ctx.MustGet("log").(*logrus.Entry)

	req := model.LoginRoomReq{}
	if err := model.Decode(ctx, &req); err != nil {
		log.Errorf("guest join room failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	userE, err := op.LoadOrInitUserByID(db.GuestUserID)
	if err != nil {
		log.Errorf("guest join room failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}
	user := userE.Value()

	roomE, err := op.LoadOrInitRoomByID(req.RoomId)
	if err != nil {
		log.Errorf("guest join room failed: %v", err)
		if err == op.ErrRoomBanned || err == op.ErrRoomPending {
			ctx.AbortWithStatusJSON(http.StatusForbidden, model.NewApiErrorResp(err))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusNotFound, model.NewApiErrorResp(err))
		return
	}
	room := roomE.Value()

	if !room.CheckPassword(req.Password) {
		log.Warn("guest join room failed: password error")
		ctx.AbortWithStatusJSON(http.StatusForbidden, model.NewApiErrorStringResp("password error"))
		return
	}

	token, err := middlewares.NewAuthRoomToken(user, room)
	if err != nil {
		log.Errorf("guest join room failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"roomId": room.ID,
		"token":  token,
	}))
}

func LoginRoom(ctx *gin.Context) {
	user := ctx.MustGet("user").(*op.UserEntry).Value()
	log := ctx.MustGet("log").(*logrus.Entry)

	req := model.LoginRoomReq{}
	if err := model.Decode(ctx, &req); err != nil {
		log.Errorf("login room failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	roomE, err := op.LoadOrInitRoomByID(req.RoomId)
	if err != nil {
		log.Errorf("login room failed: %v", err)
		if err == op.ErrRoomBanned || err == op.ErrRoomPending {
			ctx.AbortWithStatusJSON(http.StatusForbidden, model.NewApiErrorResp(err))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusNotFound, model.NewApiErrorResp(err))
		return
	}
	room := roomE.Value()

	if !user.IsAdmin() && !user.IsRoomAdmin(room) && !room.CheckPassword(req.Password) {
		log.Warn("login room failed: password error")
		ctx.AbortWithStatusJSON(http.StatusForbidden, model.NewApiErrorStringResp("password error"))
		return
	}

	token, err := middlewares.NewAuthRoomToken(user, room)
	if err != nil {
		log.Errorf("login room failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"roomId": room.ID,
		"token":  token,
	}))
}

func DeleteRoom(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.RoomEntry)
	user := ctx.MustGet("user").(*op.UserEntry).Value()
	log := ctx.MustGet("log").(*logrus.Entry)

	if err := user.DeleteRoom(room); err != nil {
		log.Errorf("delete room failed: %v", err)
		if errors.Is(err, dbModel.ErrNoPermission) {
			ctx.AbortWithStatusJSON(
				http.StatusForbidden,
				model.NewApiErrorResp(
					fmt.Errorf("delete room failed: %w", err),
				),
			)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func SetRoomPassword(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.RoomEntry).Value()
	user := ctx.MustGet("user").(*op.UserEntry).Value()
	log := ctx.MustGet("log").(*logrus.Entry)

	req := model.SetRoomPasswordReq{}
	if err := model.Decode(ctx, &req); err != nil {
		log.Errorf("set room password failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := user.SetRoomPassword(room, req.Password); err != nil {
		log.Errorf("set room password failed: %v", err)
		if errors.Is(err, dbModel.ErrNoPermission) {
			ctx.AbortWithStatusJSON(
				http.StatusForbidden,
				model.NewApiErrorResp(
					fmt.Errorf("set room password failed: %w", err),
				),
			)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	token, err := middlewares.NewAuthRoomToken(user, room)
	if err != nil {
		log.Errorf("set room password failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"roomId": room.ID,
		"token":  token,
	}))
}

func RoomSetting(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.RoomEntry).Value()
	// user := ctx.MustGet("user").(*op.UserEntry)

	ctx.JSON(http.StatusOK, model.NewApiDataResp(room.Settings))
}

func SetRoomSetting(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.RoomEntry).Value()
	user := ctx.MustGet("user").(*op.UserEntry).Value()
	log := ctx.MustGet("log").(*logrus.Entry)

	req := model.SetRoomSettingReq{}
	if err := model.Decode(ctx, &req); err != nil {
		log.Errorf("set room setting failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := user.UpdateRoomSettings(room, req); err != nil {
		log.Errorf("set room setting failed: %v", err)
		if errors.Is(err, dbModel.ErrNoPermission) {
			ctx.AbortWithStatusJSON(
				http.StatusForbidden,
				model.NewApiErrorResp(
					fmt.Errorf("set room setting failed: %w", err),
				),
			)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
