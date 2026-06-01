package re4business

import (
	"strings"
	"testing"
	"unicode"
)

// 测试能否匹配出因为重名而进行重命名的文件
// 文件名格式
// <基本名>.<原始拓展名>.1.1.1.enc
// .1这部分是为了防止重名而加的防护性措施, 重名时一般直接给.1加一变成.2, 个别情况下可能出现.1.1.1的情况
// .enc是固定的, 表示加密文件的后缀
// 要求匹配格式并把基本名和原始拓展名提取出来

// isDigitOnly 检查字符串是否只包含数字
func isDigitOnly(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// ParseEncFilename 解析加密文件名，返回基本名和原始扩展名
func ParseEncFilename(filename string) (basename string, ext string, matched bool) {
	// 必须以 .enc 结尾（大小写敏感）
	if !strings.HasSuffix(filename, ".enc") {
		return "", "", false
	}

	// 去掉 .enc 后缀
	nameWithoutEnc := filename[:len(filename)-4]

	// 按 . 分割
	parts := strings.Split(nameWithoutEnc, ".")
	if len(parts) == 0 {
		return "", "", false
	}

	// 从后往前找，跳过数字部分（防重名标记）
	i := len(parts) - 1
	for i >= 0 && isDigitOnly(parts[i]) {
		i--
	}

	// 如果全部都是数字，没有扩展名
	if i < 0 {
		return "", "", false
	}

	// 剩下的部分是 parts[0..i]
	remainingParts := parts[:i+1]

	// 如果只有一个部分，没有扩展名
	if len(remainingParts) == 1 {
		return remainingParts[0], "", true
	}

	// 基本名是第一个部分
	basename = remainingParts[0]
	// 扩展名是剩下的所有部分
	ext = strings.Join(remainingParts[1:], ".")

	return basename, ext, true
}

func TestFilenameRegex(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		wantMatch bool
		wantBase  string
		wantExt   string
	}{
		{"no enc suffix", "file.txt", false, "", ""},
		{"simple", "file.txt.enc", true, "file", "txt"},
		{"with number", "file.txt.1.enc", true, "file", "txt"},
		{"with large number", "file.txt.123456.enc", true, "file", "txt"},
		{"with multiple numbers", "file.txt.1.2.3.4.5.6.enc", true, "file", "txt"},
		{"with two numbers", "file.txt.3.4.enc", true, "file", "txt"},
		{"wrong suffix", "file.txt.xxx", false, "", ""},
		{"enc with extra", "file.txt.encxxx", false, "", ""},
		{"uppercase enc", "file.txt.ENC", false, "", ""},
		{"no original ext", "file.enc", true, "file", ""},
		{"no ext with number", "file.1.enc", true, "file", ""},
		{"nested dots", "archive.tar.gz.enc", true, "archive", "tar.gz"},
		{"nested dots with numbers", "archive.tar.gz.1.2.enc", true, "archive", "tar.gz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotBase, gotExt, gotMatch := ParseEncFilename(tc.input)
			if gotMatch != tc.wantMatch {
				t.Errorf("ParseEncFilename(%q).matched = %v, want %v", tc.input, gotMatch, tc.wantMatch)
			}
			if gotMatch {
				if gotBase != tc.wantBase {
					t.Errorf("ParseEncFilename(%q).basename = %q, want %q", tc.input, gotBase, tc.wantBase)
				}
				if gotExt != tc.wantExt {
					t.Errorf("ParseEncFilename(%q).ext = %q, want %q", tc.input, gotExt, tc.wantExt)
				}
			}
		})
	}
}
