package errs

const (
	// E 错误
	E    = 50000
	BIND = 10000
)

var (
	SystemError                = New("系统错误", E)
	InvalidToken               = New("登录信息已失效，请重新登录", 40300)
	ExpiredToken               = New("登录信息已失效，请重新登录", 40200)
	EmptyToken                 = New("请先登录", 40100)
	ParamError                 = New("参数错误", 404000)
	SystemBusy                 = New("系统繁忙，请稍后再试", 404001)
	SMSError                   = New("短信验证码输入错误", 405000)
	LoginError                 = New("用户名或密码不正确", 405001)
	ValidateMobileError        = New("请先校验手机", 405002)
	MobileBindError            = New("该手机已绑定其他手机,请先解绑", 405003)
	MobileError                = New("请输入正确的手机号", 405004)
	PermissionError            = New("您无权限进行该操作，请联系团队管理员", 405005)
	TeamUserOffError           = New("账号未启用，请联系管理员", 405006)
	TeamUserURLExpiredOffError = New("邀请链接已过期", 405007)
	AddTeamUserExistError      = New("用户已加入团队", 405008)
	AddTeamUserNoFoundError    = New("用户不存在", 405009)
	UpdateVIPError             = New("请升级会员", 405010)
	ReviewError                = New("请下载之后再预览", 405013)
	ClipAllError               = New("当前有整场下载任务未完成", 405014)
	SwitchTeamError            = New("你已切换团队，当前页面已失效，请返回首页。", 405006)
	GroupExistError            = New("分组已存在", 405007)
	GroupNoExistError          = New("分组不存在", 405008)
	AuthorNoExistError         = New("达人不存在", 405009)
	AuthorExpireError          = New("栏位已过期", 405010)
	UpdateEnterpriseError      = New("请升级企业会员", 405011)
	ClipOverTimeError          = New("不能超过1小时", 405012)
	AuthorAllotError           = New("分配达人异常,请重试", 405013)
	BindWxError                = New("该手机已经绑定其他微信", 10000)
	EndBindingOperation        = New("已取消绑定操作", 10002)
	BindWxRepeatError          = New("该微信已绑定其他用户", 10003)
	BindWxFirstError           = New("请先绑定微信", 10004)
	BindCloneError             = New("该团长已绑定其他用户", 10005)
	ProductNotFoundError       = New("获取商品信息异常", 10001)
	//	BAIYING_PRODUCT_NOT_FOUND_ERROR = New("联盟商品信息不存在", 10002)

	// 小程序
	AppletGrantNotFoundError       = New("记录不存在", 406000)
	AppletGrantError               = New("请先解绑再操作", 406001)
	AppletGrantAuthorNotFoundError = New("达人不存在", 406002)
	AppletKolNotFoundError         = New("团长不存在", 406003)
	CourseUnlockError              = New("当前课程未解锁", 406004)

	// 涉及权限的错误码以5开头
	DownloadMinuteNotEnough = New("您的下载分钟数不足，请先购买分钟数", 50001)
	SpaceNotEnough          = New("您的存储空间不足，请先删除无用数据或者购买存储", 50002)

	NotEnoughAsset = New("余额不足", 600001)
)

// HTTPException HTTP错误
type HTTPException struct {
	Message string
	Code    int
}

// Error 错误
func (h HTTPException) Error() string {
	return h.Message
}

// New 返回一个新的错误
func New(message string, codes ...int) *HTTPException {
	code := E
	if len(codes) > 0 {
		code = codes[0]
	}
	return &HTTPException{
		Message: message,
		Code:    code,
	}
}

func (h *HTTPException) Output(message string) *HTTPException {
	return New(message, h.Code)
}
