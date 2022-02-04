package verification

import "github.com/gin-gonic/gin"

var (
	X_GENZ_TOKEN = "4439EA5BDBA8B179722265789D029477"
)
	

func VerifyInternalKey(ctx *gin.Context)(bool, string) {
	var xGenzToken string
	xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"];
	if !ok {
		return false, "Access denied: token does not exist"
	}
	xGenzToken = xGenzTokenArr[0]
	if xGenzToken != X_GENZ_TOKEN {
		return false, "Access denied: invalid token."
	}

	return true, ""
}