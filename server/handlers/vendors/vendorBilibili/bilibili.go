package vendorBilibili

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"github.com/synctv-org/synctv/internal/db"
	"github.com/synctv-org/synctv/internal/op"
	"github.com/synctv-org/synctv/internal/vendor"
	"github.com/synctv-org/synctv/server/model"
	"github.com/synctv-org/synctv/utils"
	"github.com/synctv-org/vendors/api/bilibili"
)

type ParseReq struct {
	URL string `json:"url"`
}

func (r *ParseReq) Validate() error {
	if r.URL == "" {
		return errors.New("url is empty")
	}
	return nil
}

func (r *ParseReq) Decode(ctx *gin.Context) error {
	return json.NewDecoder(ctx.Request.Body).Decode(r)
}

func Parse(ctx *gin.Context) {
	user := ctx.MustGet("user").(*op.UserEntry).Value()

	req := ParseReq{}
	if err := model.Decode(ctx, &req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	var cli = vendor.LoadBilibiliClient(ctx.Query("backend"))

	resp, err := cli.Match(ctx, &bilibili.MatchReq{
		Url: req.URL,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	// can be no login
	var cookies []*http.Cookie
	bucd, err := user.BilibiliCache().Get(ctx)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound("vendor")) {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}
	} else {
		cookies = bucd.Cookies
	}

	switch resp.Type {
	case "bv":
		resp, err := cli.ParseVideoPage(ctx, &bilibili.ParseVideoPageReq{
			Cookies:  utils.HttpCookieToMap(cookies),
			Bvid:     resp.Id,
			Sections: ctx.DefaultQuery("sections", "false") == "true",
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
			return
		}
		ctx.JSON(http.StatusOK, model.NewApiDataResp(resp))
	case "av":
		aid, err := strconv.ParseUint(resp.Id, 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}
		resp, err := cli.ParseVideoPage(ctx, &bilibili.ParseVideoPageReq{
			Cookies:  utils.HttpCookieToMap(cookies),
			Aid:      aid,
			Sections: ctx.DefaultQuery("sections", "false") == "true",
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
			return
		}
		ctx.JSON(http.StatusOK, model.NewApiDataResp(resp))
	case "ep":
		epid, err := strconv.ParseUint(resp.Id, 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}
		resp, err := cli.ParsePGCPage(ctx, &bilibili.ParsePGCPageReq{
			Cookies: utils.HttpCookieToMap(cookies),
			Epid:    epid,
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
			return
		}
		ctx.JSON(http.StatusOK, model.NewApiDataResp(resp))
	case "ss":
		ssid, err := strconv.ParseUint(resp.Id, 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}
		resp, err := cli.ParsePGCPage(ctx, &bilibili.ParsePGCPageReq{
			Cookies: utils.HttpCookieToMap(cookies),
			Ssid:    ssid,
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
			return
		}
		ctx.JSON(http.StatusOK, model.NewApiDataResp(resp))
	case "live":
		roomid, err := strconv.ParseUint(resp.Id, 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}
		resp, err := cli.ParseLivePage(ctx, &bilibili.ParseLivePageReq{
			// Cookies: utils.HttpCookieToMap(cookies), // maybe no need login
			RoomID: roomid,
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
			return
		}
		ctx.JSON(http.StatusOK, model.NewApiDataResp(resp))
	default:
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorStringResp(fmt.Sprintf("unknown match type %s", resp.Type)))
		return
	}
}
