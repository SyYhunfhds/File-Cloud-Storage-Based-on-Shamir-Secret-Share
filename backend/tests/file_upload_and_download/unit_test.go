package file_upload_and_download

import (
	"archive/zip"
	sharev1 "backend/api/share/v1"
	"backend/internal/model/entity"
	"context"
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"io"
	"runtime"
	"strings"
	"testing"

	authv1 "backend/api/auth/v1"
	itemv1 "backend/api/item/v1"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/davecgh/go-spew/spew"
	"github.com/gogf/gf/v2/encoding/ghtml"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/grand"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func randUserInfo() entity.Users {
	return entity.Users{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 16),
		Email:    gofakeit.Email(),
	}
}

func testRegisterAndLogin(t *testing.T, user entity.Users) {
	ctx := gctx.New()
	// var rs string

	// 注册
	resp, err := g.Client().Post(ctx, fmt.Sprintf("%s://%s/%s", scheme, server, register), g.Map{
		"username": user.Username,
		"password": user.Password,
		"email":    user.Email,
	})
	if err != nil {
		t.Errorf("register failed: %v", err)
		t.FailNow()
		resp.Close()
		return
	}
	t.Logf("Register success with %v", user)
	// t.Logf("Register response: %v", resp.ReadAllString())
	resp.Close()

	// 登录
	resp, err = g.Client().Post(ctx, fmt.Sprintf("%s://%s/%s", scheme, server, login), g.Map{
		"username": user.Username,
		"password": user.Password,
		// "email":    user.Email,
	})
	if err != nil {
		t.Errorf("login failed: %v", err)
		resp.Close()
		t.FailNow()
		return
	}
	t.Logf("Login success with %v", user)
	// t.Logf("Login response: %v", resp.ReadAllString())

	// 序列化参数
	res := authv1.LoginRes{}
	if err = gjson.DecodeTo(resp.ReadAllString(), &res); err != nil {
		t.Errorf("decode login response failed: %v", err)
		resp.Close()
		t.FailNow()
		return
	}
	// spew.Dump(res)
	resp.Close()

	// 访问用户个人信息
	resp, err = g.Client().
		SetHeader("Authorization", res.Data).
		Get(ctx, fmt.Sprintf("%s://%s/%s", scheme, server, me))
	if resp != nil {
		// rs = resp.ReadAllString()
		resp.Close()
	}
	if err != nil {
		t.Errorf("get user info failed: %v", err)
		t.FailNow()
		return
	}
	// t.Logf("Get user info success: %v", rs)
}

func getJWTToken(ctx context.Context, user entity.Users) (token string, err error) {
	params := g.Map{
		"username": user.Username,
		"password": user.Password,
	}
	res := authv1.LoginRes{}

	r, err := g.Client().Post(ctx, fmt.Sprintf("%s://%s/%s", scheme, server, login), params)
	if err != nil {
		return "", err
	}

	if err = gjson.DecodeTo(r.ReadAllString(), &res); err != nil {
		return "", err
	}
	return res.Data, nil
}

func getClient(token string) func() *gclient.Client {
	return func() *gclient.Client {
		return g.Client().SetHeader("Authorization", token)
	}
}

type fileGenFunc func(size ...int) (filename string, content []byte, n int)

func testUploadAnyFile(t *testing.T, token string, generate fileGenFunc) {
	filename, content, n := generate()
	if n == 0 {
		t.Errorf("生成文件失败")
		return
	}
	originalCrc := crc32.ChecksumIEEE(content)
	t.Logf("上传文件: %s, 大小: %s", gfile.Basename(filename), gfile.FormatSize(int64(n)))
	cleanup, path, _ := saveTempFile(filename, content)
	defer cleanup()
	defer runtime.GC() // 手动触发GC

	var resp *gclient.Response
	var err error
	var rs string
	resp, err = g.Client().
		SetHeader("Authorization", token).
		Post(gctx.New(), fmt.Sprintf("%s://%s/%s", scheme, server, upload),
			"item=@file:"+path,
		)
	if resp != nil {
		rs = resp.ReadAllString()
		resp.Close()
	}
	if err != nil {
		t.Errorf("upload file failed: %v", err)
		return
	}
	// t.Logf("上传HTML文件响应: %v", rs)

	var uploadRes = itemv1.ItemSubmitRes{}
	if err = gjson.DecodeTo(rs, &uploadRes); err != nil {
		t.Errorf("decode upload response failed: %v", err)
		return
	}
	// spew.Dump(uploadRes)

	var (
		itemId      = uploadRes.Data.ItemId
		deviceShare = uploadRes.Data.Share
	)

	// 文件下载
	var downloadedFile []byte
	resp, err = g.Client().SetHeader("Authorization", token).
		Post(gctx.New(), fmt.Sprintf("%s://%s/%s", scheme, server, download), g.Map{
			"item_id": itemId,
			"share":   deviceShare,
		})
	if resp != nil {
		downloadedFile = resp.ReadAll()
		resp.Close()
	}
	if err != nil {
		t.Errorf("download file failed: %v", err)
		return
	}

	if len(downloadedFile) != len(content) {
		fmt.Printf("下载文件内容: \n%s\n", string(downloadedFile))
		t.Errorf("下载文件大小不一致, 原: %v, 下载后: %v",
			gfile.FormatSize(int64(len(content))), gfile.FormatSize(int64(len(downloadedFile))))
	}

	currCrc32 := crc32.ChecksumIEEE(downloadedFile)
	t.Logf("上传前文件校验和为 %d, 下载后文件校验和为 %d", originalCrc, currCrc32)
	if currCrc32 != originalCrc {
		t.Errorf("文件CRC32校验失败, 原: %v, 下载后: %v", originalCrc, currCrc32)
		return
	}
}

func TestUploadAndDownload(t *testing.T) {
	ctx := gctx.New()
	// 测试上传和下载
	user := randUserInfo()
	testRegisterAndLogin(t, user) // 先登录看看行不行
	if t.Failed() {
		t.Skip()
	}

	token, err := getJWTToken(ctx, user)
	if err != nil {
		t.Errorf("get token failed: %v", err)
		t.FailNow()
		return
	}

	testUploadAnyFile(t, token, randHTMLFile)
	// testUploadAnyFile(t, token, randImageFile)
	testUploadAnyFile(t, token, randZipFile)
}

// randHTMLFile 接受一个参数用于指定HTML内随机纯文本的长度
func randHTMLFile(size ...int) (filename string, content []byte, contentLen int) {
	var filesize int
	if len(size) > 0 && size[0] > 0 {
		filesize = size[0]
	} else {
		filesize = grand.Intn(8 * MB)
	}

	htmltext := ghtml.Entities(grand.S(filesize))
	filename = fmt.Sprintf("%s.html", grand.S(6))

	return filename, []byte(htmltext), len(htmltext)
}

// randImageFile 不接受任何参数, 固定获取1080P图片
func randImageFile(size ...int) (filename string, content []byte, contentLen int) {
	resp, err := g.Client().Get(gctx.New(), "https://picsum.photos/1920/1080") // 随机1080P图片
	if err != nil {
		return
	}
	filename = fmt.Sprintf("%s.jpg", grand.S(6))
	content = resp.ReadAll()

	return filename, content, len(content)
}

func randZipFile(size ...int) (filename string, content []byte, contentLen int) {
	zipFilename := grand.S(6) + ".zip"
	zipFile, err := gfile.Create(zipFilename)
	if err != nil {
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	contentWriter, err := zipWriter.Create(grand.S(6) + ".bin")
	if err != nil {
		return
	}

	var filesize int
	if len(size) > 0 && size[0] > 0 {
		filesize = size[0]
	} else {
		filesize = grand.Intn(8 * MB)
	}
	randomSource := io.LimitReader(rand.Reader, int64(filesize))
	_, err = io.Copy(contentWriter, randomSource)
	if err != nil {
		return
	}

	content = gfile.GetBytes(zipFilename)
	_ = gfile.RemoveFile(zipFilename)

	return zipFilename, content, len(content)
}

func saveTempFile(filename string, file []byte) (cleanup func(), path string, err error) {
	dir := gfile.Temp(grand.S(16))
	path = gfile.Join(dir, filename)

	err = gfile.PutBytes(path, file)
	cleanup = func() {
		gfile.RemoveAll(dir)
	}
	return
}

func getContentCRC32(content []byte) (uint32, error) {
	hasher := crc32.NewIEEE()
	if _, err := hasher.Write(content); err != nil {
		return 0, err
	}
	return hasher.Sum32(), nil
}

type GenericResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

var (
	registerAPI = fmt.Sprintf("%s://%s/%s", scheme, server, register)
	loginAPI    = fmt.Sprintf("%s://%s/%s", scheme, server, login)

	uploadAPI     = fmt.Sprintf("%s://%s/%s", scheme, server, upload)
	downloadAPI   = fmt.Sprintf("%s://%s/%s", scheme, server, download)
	updateAPI     = fmt.Sprintf("%s://%s/%s", scheme, server, update)
	viewSubmitAPI = fmt.Sprintf("%s://%s/%s", scheme, server, view)
	passAllAPI    = fmt.Sprintf("%s://%s/%s", scheme, server, passAll)

	refreshAPI = fmt.Sprintf("%s://%s/%s", scheme, server, refresh)
)

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
func last[T any](s []T) T {
	return s[len(s)-1]
}
