package vendorAlist

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"github.com/synctv-org/synctv/internal/cache"
	"github.com/synctv-org/synctv/internal/db"
	dbModel "github.com/synctv-org/synctv/internal/model"
	"github.com/synctv-org/synctv/internal/op"
	"github.com/synctv-org/synctv/server/model"
)

type LoginReq struct {
	Host           string `json:"host"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	HashedPassword string `json:"hashedPassword"`
}

func (r *LoginReq) Validate() error {
	if r.Host == "" {
		return errors.New("host is required")
	}
	if r.Password != "" && r.HashedPassword != "" {
		return errors.New("password and hashedPassword can't be both set")
	}
	return nil
}

func (r *LoginReq) Decode(ctx *gin.Context) error {
	return json.NewDecoder(ctx.Request.Body).Decode(r)
}

func Login(ctx *gin.Context) {
	user := ctx.MustGet("user").(*op.User)

	req := LoginReq{}
	if err := model.Decode(ctx, &req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if req.Password != "" {
		h := sha256.New()
		h.Write([]byte(req.Password + `-https://github.com/alist-org/alist`))
		req.HashedPassword = hex.EncodeToString(h.Sum(nil))
	}

	backend := ctx.Query("backend")

	data, err := cache.AlistAuthorizationCacheWithConfigInitFunc(ctx, &dbModel.AlistVendor{
		Host:           req.Host,
		Username:       req.Username,
		HashedPassword: []byte(req.HashedPassword),
		Backend:        backend,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	_, err = db.CreateOrSaveAlistVendor(user.ID, &dbModel.AlistVendor{
		Backend:        backend,
		Host:           req.Host,
		Username:       req.Username,
		HashedPassword: []byte(req.HashedPassword),
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	_, err = user.AlistCache().Data().Refresh(ctx, func(ctx context.Context, args ...struct{}) (*cache.AlistUserCacheData, error) {
		return data, nil
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func Logout(ctx *gin.Context) {
	user := ctx.MustGet("user").(*op.User)

	err := db.DeleteAlistVendor(user.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	user.AlistCache().Clear()

	ctx.Status(http.StatusNoContent)
}
