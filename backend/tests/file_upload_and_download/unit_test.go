package file_upload_and_download

import (
	"archive/zip"
	"backend/internal/model/entity"
	"context"
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"io"
	"runtime"
	"testing"

	authv1 "backend/api/auth/v1"
	itemv1 "backend/api/item/v1"

	"github.com/brianvoe/gofakeit/v7"
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
	t.Logf("上传HTML文件响应: %v", rs)

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
