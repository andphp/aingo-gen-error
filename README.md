## go-error-generator
> go-error-generator是一个通过protobuf文件的Enum对象自动生成Error的插件，通过在扩展的EnumValueOptions中定义多个option轻松实现error的i18n
> https://github.com/classtorch/go-error-generator/blob/master/README_zh.md

## 错误码设计规范
#### 错误码为 6 位数

| 1 | 01 | 001   |
| :------ | :------ |:------|
| 服务级错误码 | 模块级错误码 | 具体错误码 |

- 服务级错误码：1 位数进行表示，比如 1 为系统级错误；2 为普通错误，通常是由用户非法操作引起。
- 模块级错误码：2 位数进行表示，比如 01 为用户模块；02 为订单模块。
- 具体的错误码：3 位数

```golang
type ErrorCode int32
```

### 生成ErrorCode示例
```golang
const (
	// System Level Errors
	ErrorCode_SUCCESS         ErrorCode = 0
	ErrorCode_UNKNOWN         ErrorCode = 100000
	ErrorCode_VALIDATION_FAIL ErrorCode = 100001
	ErrorCode_LOGIN_ERROR     ErrorCode = 100002
	ErrorCode_ACCESS_DENIED   ErrorCode = 100003
	ErrorCode_API_AUTH_FAIL   ErrorCode = 100004
	// API Parameter Validation Errors (4xxxxx)
	ErrorCode_INVALID_PARAMS                 ErrorCode = 400001
	ErrorCode_NO_TOKEN                       ErrorCode = 400002
	ErrorCode_CODE_ERROR                     ErrorCode = 400003
	ErrorCode_NOT_ALLOW_METHOD               ErrorCode = 400004
	ErrorCode_UPLOAD_IMAGE_VERIFICATION_FAIL ErrorCode = 400005
	ErrorCode_UPLOAD_FILE_VERIFICATION_FAIL  ErrorCode = 400006
	// Business Logic Validation Errors (5xxxxx)
	ErrorCode_LOGIN_FAILURE         ErrorCode = 500001
	ErrorCode_USER_BAN              ErrorCode = 500002
	ErrorCode_NO_PERMISSION         ErrorCode = 500003
	ErrorCode_DIRECTORY_NO_DELETE   ErrorCode = 500004
	ErrorCode_QUEUE_MISSING_CONTENT ErrorCode = 500005
	ErrorCode_API_STOP              ErrorCode = 500006
)
```

### 生成变量示例，Msg为默认，多语言配置文件定义给Msg加后缀 Msg_English
```golang
var (
	Msg = map[int32]*errors.Error{
		0:      &errors.Error{Code: 0, Msg: "请求成功"},
		100000: &errors.Error{Code: 100000, Msg: "未知错误"},
		100001: &errors.Error{Code: 100001, Msg: "Token 验证失败，可能已过期或者在黑名单"},
		100002: &errors.Error{Code: 100002, Msg: "用户不存在或密码不正确"},
		100003: &errors.Error{Code: 100003, Msg: "拒绝非法访问"},
		100004: &errors.Error{Code: 100004, Msg: "接口鉴权失败"},
		400001: &errors.Error{Code: 400001, Msg: "请求参数错误"},
		400002: &errors.Error{Code: 400002, Msg: "Token 不合法或者不存在"},
		400003: &errors.Error{Code: 400003, Msg: "验证码错误或已失效"},
		400004: &errors.Error{Code: 400004, Msg: "资源未找到或不允许访问的资源"},
		400005: &errors.Error{Code: 400005, Msg: "图片上传验证不通过"},
		400006: &errors.Error{Code: 400006, Msg: "文件上传验证不通过"},
		500001: &errors.Error{Code: 500001, Msg: "用户名不存在或者密码错误!"},
		500002: &errors.Error{Code: 500002, Msg: "用户被禁用"},
		500003: &errors.Error{Code: 500003, Msg: "无权限访问"},
		500004: &errors.Error{Code: 500004, Msg: "非空目录，不可删除"},
		500005: &errors.Error{Code: 500005, Msg: "队列消息缺少消息内容"},
		500006: &errors.Error{Code: 500006, Msg: "接口已停用"},
	}

	Msg_English = map[int32]*errors.Error{
		0:      &errors.Error{Code: 0, Msg: "request successful"},
		100000: &errors.Error{Code: 100000, Msg: "unknown error"},
		100001: &errors.Error{Code: 100001, Msg: "token validation failed, might be expired or blacklisted"},
		100002: &errors.Error{Code: 100002, Msg: "user does not exist or password incorrect"},
		100003: &errors.Error{Code: 100003, Msg: "access denied"},
		100004: &errors.Error{Code: 100004, Msg: "API authentication failed"},
		400001: &errors.Error{Code: 400001, Msg: "request parameter error"},
		400002: &errors.Error{Code: 400002, Msg: "token is invalid or does not exist"},
		400003: &errors.Error{Code: 400003, Msg: "verification code error or expired"},
		400004: &errors.Error{Code: 400004, Msg: "method [:method] not allowed for accessing resource"},
		400005: &errors.Error{Code: 400005, Msg: "image upload verification failed"},
		400006: &errors.Error{Code: 400006, Msg: "file upload verification failed"},
		500001: &errors.Error{Code: 500001, Msg: "The username does not exist or the password is incorrect"},
		500002: &errors.Error{Code: 500002, Msg: "user is banned"},
		500003: &errors.Error{Code: 500003, Msg: "no permission to access"},
		500004: &errors.Error{Code: 500004, Msg: "non-empty directory, cannot delete"},
		500005: &errors.Error{Code: 500005, Msg: "queue message missing content"},
		500006: &errors.Error{Code: 500006, Msg: "API has been discontinued"},
	}
)
```

> 读取当前目录下config.json 配置文件


```json
{
  "serviceCodes": [
    {
      "code": "1",
      "label":"系统级错误",
      "desc":"1 为系统级错误"
    },
    {
      "code": "2",
      "label":"普通错误",
      "desc":"2 为普通错误，通常是由用户非法操作引起"
    }
  ],
  "moduleCodes": [
    {
      "code": "01",
      "label":"用户模块",
      "desc":"01 为用户模块"
    },
    {
      "code": "02",
      "label":"订单模块",
      "desc":"02 为订单模块"
    }
  ],
  "file_path": "./errors.proto",
  "i18n": ["default","English"]
}
```



## 待生成的示例模板

```golang
syntax = "proto3";

package errcode;

option go_package = "err/errcode";
import "errors/errors.proto";

enum ErrorCode {
// System Level Errors
UNKNOWN = 100000 [(errors.msg) = "未知错误", (errors.msg_english) = "unknown error"];

// Business Logic Validation Errors (5xxxxx)
INVALID_PARAMS = 200001 [(errors.msg) = "请求参数错误", (errors.msg_english) = "request parameter error"];
NO_TOKEN = 200002 [(errors.msg) = "Token 不合法或者不存在", (errors.msg_english) = "token is invalid or does not exist"];

LOGIN_FAILURE = 201001 [(errors.msg) = "用户名不存在或者密码错误!", (errors.msg_english) = "The username does not exist or the password is incorrect"];
USER_BAN = 201002 [(errors.msg) = "用户被禁用", (errors.msg_english) = "user is banned"];

}
```

> 执行命令 aingo-gen.exe error