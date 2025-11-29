package admin

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"time"

	"github.com/goal-web/application"
	"github.com/goal-web/contracts"

	"github.com/qbhy/goal-admin/models"
	adminreq "github.com/qbhy/goal-admin/requests/admin"
	adminres "github.com/qbhy/goal-admin/results/admin"
)

func init() {
	AuthServiceDefine.Login = Login
	AuthServiceDefine.SendSms = SendSms
	AuthServiceDefine.GetCaptcha = GetCaptcha
	AuthServiceDefine.ResetPassword = ResetPassword
	AuthServiceDefine.GetCurrentUser = GetCurrentUser
}

// Login 管理员登录（支持：手机号+验证码 或 账号+密码）
func Login(req *adminreq.LoginReq, ctx contracts.Context) (*adminres.LoginResult, error) {
	// 短信验证码登录
	if req.Code != "" {
		if !isValidPhone(req.Phone) {
			return nil, fmt.Errorf("手机号格式不正确")
		}
		if !validateSmsCode(req.Phone, "login", req.Code) {
			return nil, fmt.Errorf("验证码错误或已过期")
		}
		adminInstance, err := models.AdminQuery().Where("phone", req.Phone).FirstE()
		if err != nil {
			return nil, fmt.Errorf("管理员不存在")
		}
		guard := application.Get("auth").(contracts.Auth).Guard("admin_jwt", ctx)
		token, ok := guard.Login(adminInstance).(string)
		if !ok {
			return nil, fmt.Errorf("登录失败")
		}
		return &adminres.LoginResult{Token: token, Admin: *adminInstance}, nil
	}

	// 账号密码登录
	if req.Username == "" || req.Password == "" {
		return nil, fmt.Errorf("用户名或密码不能为空")
	}

	adminInstance, err := models.AdminQuery().Where("username", req.Username).FirstE()
	if err != nil {
		return nil, fmt.Errorf("管理员不存在")
	}

	hashing := application.Get("hash").(contracts.Hasher)
	if !hashing.Check(req.Password, adminInstance.Password, contracts.Fields{}) {
		return nil, fmt.Errorf("密码错误")
	}

	guard := application.Get("auth").(contracts.Auth).Guard("admin_jwt", ctx)
	token, ok := guard.Login(adminInstance).(string)
	if !ok {
		return nil, fmt.Errorf("登录失败")
	}
	return &adminres.LoginResult{Token: token, Admin: *adminInstance}, nil
}

// SendSms 发送管理员短信验证码
func SendSms(req *adminreq.SendSmsReq, ctx contracts.Context) (*adminres.SendSmsResult, error) {
	return &adminres.SendSmsResult{Success: true, Message: "发送成功"}, nil
}

// GetCaptcha 生成图形验证码
func GetCaptcha(req *adminreq.GetCaptchaReq, ctx contracts.Context) (*adminres.GetCaptchaResult, error) {
	cacheStore := application.Get("cache.store").(contracts.CacheStore)

	code := generateCaptchaCode(5)
	id := randomCaptchaID()

	cacheStore.Put(fmt.Sprintf("captcha:%s", id), code, 2*time.Minute)

	svg := fmt.Sprintf(`<svg xmlns='http://www.w3.org/2000/svg' width='160' height='48'>
<rect width='100%%' height='100%%' fill='#fff'/>
<text x='50%%' y='50%%' dominant-baseline='middle' text-anchor='middle' font-family='monospace' font-size='28' fill='#000'>%s</text>
</svg>`, code)
	dataUri := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))

	return &adminres.GetCaptchaResult{CaptchaId: id, ImageBase64: dataUri}, nil
}

// ResetPassword 管理员忘记密码重置
func ResetPassword(req *adminreq.ResetPasswordReq, ctx contracts.Context) (*adminres.ResetPasswordResult, error) {
	if !isValidPhone(req.Phone) {
		return &adminres.ResetPasswordResult{Success: false, Message: "手机号格式不正确"}, nil
	}
	if !validateSmsCode(req.Phone, "reset", req.Code) {
		return &adminres.ResetPasswordResult{Success: false, Message: "验证码错误或已过期"}, nil
	}
	if !isValidPasswordStrength(req.NewPassword) {
		return &adminres.ResetPasswordResult{Success: false, Message: "密码强度不够，至少8位包含字母和数字"}, nil
	}

	adminInstance, err := models.AdminQuery().Where("phone", req.Phone).FirstE()
	if err != nil {
		return &adminres.ResetPasswordResult{Success: false, Message: "管理员不存在"}, nil
	}

	hashing := application.Get("hash").(contracts.Hasher)
	// 直接使用模型的 Update 方法更新字段
	hashed := hashing.Make(req.NewPassword, contracts.Fields{})
	if err := adminInstance.Update(contracts.Fields{"password": hashed}); err != nil {
		return &adminres.ResetPasswordResult{Success: false, Message: "重置失败"}, nil
	}
	return &adminres.ResetPasswordResult{Success: true, Message: "重置成功"}, nil
}

// ===== tool funcs（本文件内最小复用实现，避免跨包依赖） =====
func isValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

func isValidPasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	return hasLetter && hasNumber
}

func validateSmsCode(phone, smsType, code string) bool {
	return false
}

func generateCaptchaCode(n int) string {
	alphabet := []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = alphabet[randInt(len(alphabet))]
	}
	return string(b)
}

func randomCaptchaID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func randInt(n int) int {
	var buf [1]byte
	_, _ = rand.Read(buf[:])
	return int(buf[0]) % n
}

// GetCurrentUser 获取当前已登录的管理员信息
func GetCurrentUser(req *adminreq.GetCurrentUserReq, ctx contracts.Context) (*adminres.GetCurrentUserResult, error) {
	guard := application.Get("auth").(contracts.Auth).Guard("admin_jwt", ctx)
	user := guard.User()
	if user == nil {
		return nil, fmt.Errorf("未登录")
	}

	adminModel, ok := user.(*models.AdminModel)
	if !ok {
		return nil, fmt.Errorf("未登录")
	}

	return &adminres.GetCurrentUserResult{Admin: *adminModel}, nil
}
