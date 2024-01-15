package vendorAlist

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/synctv-org/synctv/internal/db"
	"github.com/synctv-org/synctv/internal/op"
	"github.com/synctv-org/synctv/internal/vendor"
	"github.com/synctv-org/synctv/server/model"
	"github.com/synctv-org/vendors/api/alist"
)

type AlistMeResp = model.VendorMeResp[*alist.MeResp]

func Me(ctx *gin.Context) {
	user := ctx.MustGet("user").(*op.UserEntry).Value()

	serverID := ctx.Query("serverID")
	if serverID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(errors.New("serverID is required")))
		return

	}

	aucd, err := user.AlistCache().LoadOrStore(ctx, serverID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound("vendor")) {
			ctx.JSON(http.StatusBadRequest, model.NewApiErrorStringResp("alist server not found"))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	resp, err := vendor.LoadAlistClient(aucd.Backend).Me(ctx, &alist.MeReq{
		Host:  aucd.Host,
		Token: aucd.Token,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(&AlistMeResp{
		IsLogin: true,
		Info:    resp,
	}))
}

type AlistBindsResp []*struct {
	ServerID string `json:"serverID"`
	Host     string `json:"host"`
}

func Binds(ctx *gin.Context) {
	user := ctx.MustGet("user").(*op.UserEntry).Value()

	ev, err := db.GetAlistVendors(user.ID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound("vendor")) {
			ctx.JSON(http.StatusOK, model.NewApiDataResp(&AlistMeResp{
				IsLogin: false,
			}))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	var resp AlistBindsResp = make(AlistBindsResp, len(ev))
	for i, v := range ev {
		resp[i] = &struct {
			ServerID string "json:\"serverID\""
			Host     string "json:\"host\""
		}{
			ServerID: v.ServerID,
			Host:     v.Host,
		}
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(resp))
}
