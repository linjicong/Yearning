package roles

import (
	"Yearning-go/src/handler/common"
	"Yearning-go/src/i18n"
	"github.com/cookieY/yee"
	"net/http"
)

func RolesApis() yee.RestfulAPI {
	return yee.RestfulAPI{
		Post: SuperCustomRoles,
	}
}

func SuperCustomRoles(c yee.Context) (err error) {
	tp := c.Params("tp")
	switch tp {
	case "global":
		return SuperFetchRoles(c)
	case "global_updated":
		return SuperSaveRoles(c)
	case "list":
		return SuperRolesList(c)
	case "add":
		return SuperRolesAdd(c)
	case "updated":
		return SuperRoleUpdate(c)
	case "delete":
		return SuperRoleDelete(c)
	case "profile":
		return SuperRoleProfile(c)
	}
	return c.JSON(http.StatusOK, common.ERR_COMMON_TEXT_MESSAGE(i18n.DefaultLang.Load(i18n.ER_REQ_FAKE)))

}
