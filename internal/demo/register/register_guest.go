package register

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	http2 "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

const DefaultPassword = "guest123"
const DefaultAreaCode = "86"

type ParamRegisterGuest struct {
	Nickname    string `json:"nickname"`
	CartId      string `json:"cartId"`
	Platform    int32  `json:"platform" binding:"required,min=1,max=7"`
	Ex          string `json:"ex"`
	FaceURL     string `json:"faceURL"`
	OperationID string `json:"operationID" binding:"required"`
	// AreaCode         string `json:"areaCode"`
	Password string `json:"password"`
}

func RegisterGuest(c *gin.Context) {
	params := ParamRegisterGuest{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var account string
	if params.CartId != "" {
		account = params.CartId
	}
	if params.Nickname == "" {
		params.Nickname = "访客" + account
	}

	userID := utils.Md5(params.OperationID + strconv.FormatInt(time.Now().UnixNano(), 10))
	bi := big.NewInt(0)
	bi.SetString(userID[0:8], 16)
	userID = bi.String()

	if params.Password == "" {
		params.Password = utils.Md5(DefaultPassword)
	}

	url := config.Config.Demo.ImAPIURL + "/auth/user_register"
	openIMRegisterReq := api.UserRegisterReq{}
	openIMRegisterReq.OperationID = params.OperationID
	openIMRegisterReq.Platform = params.Platform
	openIMRegisterReq.UserID = userID
	openIMRegisterReq.Nickname = params.Nickname
	openIMRegisterReq.Secret = config.Config.Secret
	openIMRegisterReq.FaceURL = params.FaceURL
	openIMRegisterResp := api.UserRegisterResp{}
	bMsg, err := http2.Post(url, openIMRegisterReq, 2)

	if err != nil {
		log.NewError(params.OperationID, "request openIM register error", account, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": err.Error()})
		return
	}

	err = json.Unmarshal(bMsg, &openIMRegisterResp)
	if err != nil || openIMRegisterResp.ErrCode != 0 {
		log.NewError(params.OperationID, "request openIM register error", account, "err", "resp: ", openIMRegisterResp.ErrCode)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), err.Error())
		}
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": "register failed: " + openIMRegisterResp.ErrMsg})
		return
	}

	log.Info(params.OperationID, "begin store mysql", account, params.Password, "info", params.FaceURL, params.Nickname)

	err = im_mysql_model.SetPassword(account, params.Password, params.Ex, userID, DefaultAreaCode)

	if err != nil {
		log.NewError(params.OperationID, "set phone number password error", account, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": err.Error()})
		return
	}
	log.Info(params.OperationID, "end setPassword", account, params.Password)

	// demo onboarding
	// onboardingProcess(params.OperationID, userID, params.Nickname, params.FaceURL, DefaultAreaCode +params.PhoneNumber, params.Email)

	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": openIMRegisterResp.UserToken})
	return
}
