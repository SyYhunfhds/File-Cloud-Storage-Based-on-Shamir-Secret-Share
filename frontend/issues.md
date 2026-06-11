## 须知
- 以下测试结果均在Windows热重载模式下测出
- 后端有Jaeger链路追踪, 所以用户说请求成功了就是成功了, 失败了就是失败了


## API对接问题
1. 份额刷新API 前端对接异常, 但调试控制台没有打印关于请求处理的任何信息(更没有报错), 后端Jaeger链路追踪也显示请求完全成功

- 附上 后端的业务测试
1. 做测试时使用的审计确认API `passAllAPI`是一个仅供调试的API, 你不需要知道这个API的业务逻辑
2. 测试代码里对SSE流式响应逐行进行序列化是不可取的, 但你可以从中知道响应行的分隔符其实是两个换行符`\n\n`
```go
func registerAndLogin(ctx context.Context, u entity.Users) (res authv1.LoginRes, err error) {
	var (
		gr = GenericResponse{}
		rs string
		r  *gclient.Response
	)
	// 注册用户
	r, err = g.Client().Post(ctx, registerAPI, g.Map{
		"username": u.Username,
		"password": u.Password,
		"email":    u.Email,
	})
	if err != nil {
		return
	}
	if r != nil {
		_ = gjson.DecodeTo(r.ReadAll(), &gr)
		if gr.Code != 0 {
			err = fmt.Errorf("register failed: %s", gr.Message)
			return res, err
		}
		_ = r.Close()
	}

	// 登录用户
	r, err = g.Client().Post(ctx, loginAPI, g.Map{
		"username": u.Username,
		"password": u.Password,
	})
	if err != nil {
		return
	}
	if r != nil {
		rs = r.ReadAllString()
		_ = r.Close()
	}
	_ = gjson.DecodeTo(rs, &res)
	if res.Code != 0 {
		err = fmt.Errorf("login failed: %s", res.Message)
		return res, err
	}
	_ = gjson.DecodeTo(rs, &res)

	return res, nil
}
func uploadFile(ctx context.Context, claim authv1.LoginRes, generate fileGenFunc) (res itemv1.ItemSubmitRes, checksum uint32, err error) {
	filename, content, _ := generate()
	cleanup, path, _ := saveTempFile(filename, content)
	checksum = crc32.ChecksumIEEE(content)
	defer cleanup()

	var (
		r  *gclient.Response
		rs string
		gr GenericResponse
	)
	r, err = g.Client().
		SetHeader("Authorization", claim.Data).
		Post(ctx, uploadAPI, "item=@file:"+path)
	if err != nil {
		return res, 0, err
	}
	if r != nil {
		rs = r.ReadAllString()
		_ = r.Close()
	}
	_ = gjson.DecodeTo(r.ReadAll(), &gr)
	if gr.Code != 0 {
		err = fmt.Errorf("upload failed: %s", gr.Message)
		return res, 0, err
	}
	_ = gjson.DecodeTo(rs, &res)

	return res, checksum, nil
}
func shareRefresh(ctx context.Context, claim authv1.LoginRes, privacy itemv1.ItemSubmitRes) (res sharev1.ShareRefreshRes, err error) {
	var (
		r  *gclient.Response
		rs []string
		sr StreamProgress
	)
	r, err = g.Client().
		SetHeader("Authorization", claim.Data).
		Post(ctx, refreshAPI, g.Map{
			"item_id":       privacy.Data.ItemId,
			"recovery_code": privacy.Data.RecoveryCode,
			"share":         privacy.Data.Share,
		})
	if err != nil {
		return res, err
	}
	if r != nil {
		rs = strings.Split(r.ReadAllString(), "\n\n")
		_ = r.Close()
	}
	for _, s := range rs {
		_ = gjson.DecodeTo(s, &sr)
		// spew.Dump(sr)
	}
	if sr.Progress != 100 {
		return res, fmt.Errorf("refresh failed: %s", sr.Message)
	}

	return sr.Payload, nil
}

type StreamProgress struct {
	Progress int                     `json:"progress"`
	Message  string                  `json:"message"`
	Payload  sharev1.ShareRefreshRes `json:"data"`
}

// 测试单用户份额强制刷新
// 直接刷新没有问题, 因此API检测不到有需要更新的审计用户就不会真正更新份额
// 有额外用户加入后文件下载依然正常, 并且Auth Share份额确实不一样
func TestShareRefresh(t *testing.T) {
	var (
		u1 = randUserInfo()

		gr = GenericResponse{}
		// rs      []string
		oners   string
		content []byte
		r       *gclient.Response
		// sr      = StreamProgress{}
		err error
	)
	_ = gr
	_ = oners
	// 登录
	claim1, err := registerAndLogin(gctx.New(), u1)
	if err != nil {
		t.Errorf("[%s@%s]登录失败", u1.Username, u1.Password)
	}

	// (用户1)上传文件拿到至少一个份额和code
	privacy, checksum, err := uploadFile(gctx.New(), claim1, randHTMLFile)
	if err != nil {
		t.Errorf("[%s@%s]文件上传失败", u1.Username, u1.Password)
		return
	}
	// (用户1)把这个条目改成公开的
	r, _ = g.Client().SetHeader("Authorization", claim1.Data).
		Post(gctx.New(), updateAPI, g.Map{
			"item_id":           privacy.Data.ItemId,
			"new_filename":      grand.S(6) + ".txt",
			"minimum_privilege": 1,
			"enable_public":     true,
		})
	if r != nil {
		oners = r.ReadAllString()
		_ = r.Close()
	}
	// println(oners)

	// (用户2)注册并登录, 申请查看用户1的条目
	var (
		u2        = randUserInfo()
		claim2, _ = registerAndLogin(gctx.New(), u2)
	)
	r, _ = g.Client().SetHeader("Authorization", claim2.Data).
		Post(gctx.New(), viewSubmitAPI, g.Map{
			"requirements": g.List{
				g.Map{
					"item_id":   privacy.Data.ItemId,
					"operation": "view",
				},
			},
		})
	if r != nil {
		oners = r.ReadAllString()
		_ = r.Close()
	}
	// println(oners)

	// (用户1) 确认用户2的申请/全部确认
	r, _ = g.Client().SetHeader("Authorization", claim1.Data).
		Get(gctx.New(), passAllAPI, nil)
	if r != nil {
		oners = r.ReadAllString()
		_ = r.Close()
	}

	// (用户1) 刷新份额
	var refreshRes sharev1.ShareRefreshRes
	refreshRes, err = shareRefresh(gctx.New(), claim1, privacy)
	if err != nil {
		t.Errorf("[%s@%s]份额刷新失败: %v", u1.Username, u1.Password, err)
		return
	}
	if refreshRes.DeviceShare == privacy.Data.Share {
		t.Errorf("份额刷新之后应该得到不一样的份额")
		return
	}

	privacy.Data.Share = refreshRes.DeviceShare
	privacy.Data.RecoveryCode = refreshRes.RecoveryCode

	// (用户1)下载文件
	r, err = g.Client().SetHeader("Authorization", claim1.Data).
		Post(gctx.New(), downloadAPI, g.Map{
			"item_id": privacy.Data.ItemId,
			"share":   refreshRes.DeviceShare,
		})
	if err != nil {
		t.Errorf("[%s@%s]文件下载失败", u1.Username, u1.Password)
		return
	}
	if r != nil {
		content = r.ReadAll()
		r.Close()
	}
	if checksum != crc32.ChecksumIEEE(content) {
		t.Errorf("[%s@%s]文件下载失败: 内容校验不匹配", u1.Username, u1.Password)
		spew.Dump(content)
		return
	}
	t.Logf("[%s@%s]文件下载成功", u1.Username, u1.Password)
}
```


