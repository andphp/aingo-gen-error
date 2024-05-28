# aingo-gen-error
golang错误码的命令行生成工具


## 错误码设计规范
#### 错误码为 6 位数

| 1 | 01 | 001   |
| :------ | :------ |:------|
| 服务级错误码 | 模块级错误码 | 具体错误码 |

- 服务级错误码：1 位数进行表示，比如 1 为系统级错误；2 为普通错误，通常是由用户非法操作引起。
- 模块级错误码：2 位数进行表示，比如 01 为用户模块；02 为订单模块。
- 具体的错误码：3 位数

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
  VALIDATION_FAIL = 100001 [(errors.msg) = "Token 验证失败，可能已过期或者在黑名单", (errors.msg_english) = "token validation failed, might be expired or blacklisted"];

  // API Parameter Validation Errors (4xxxxx)
  INVALID_PARAMS = 400001 [(errors.msg) = "请求参数错误", (errors.msg_english) = "request parameter error"];
  NO_TOKEN = 400002 [(errors.msg) = "Token 不合法或者不存在", (errors.msg_english) = "token is invalid or does not exist"];

  // Business Logic Validation Errors (5xxxxx)
  LOGIN_FAILURE = 500001 [(errors.msg) = "用户名不存在或者密码错误!", (errors.msg_english) = "The username does not exist or the password is incorrect"];
  USER_BAN = 500002 [(errors.msg) = "用户被禁用", (errors.msg_english) = "user is banned"];

}
```
