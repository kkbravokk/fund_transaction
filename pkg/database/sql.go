package database

import (
	"bytes"
	"strings"
)

// QuoteSQLOrder 针对order字段过滤
// 先按','分割后，每个数据根据空格分割数据后使用 QuoteSQLName 转义再拼接，DESC或ASC不转义，数字表示的列不转义
func QuoteSQLOrder(sort string) string {
	sortSlice := strings.Split(sort, ",")

	var escapedSlice []string
	for _, column := range sortSlice {
		col := strings.TrimSpace(column)
		if strings.HasPrefix(col, "-") { // expression
			col = col[1:] + "DESC"
		}

		fields := strings.Fields(col)
		escaped := make([]string, len(fields))
		for i, field := range fields {
			if Allow(strings.ToUpper(field), "DESC", "ASC") || isDigit(field) {
				escaped[i] = field
			} else {
				escaped[i] = QuoteSQLName(field)
			}
		}
		escapedSlice = append(escapedSlice, strings.Join(escaped, " "))
	}
	return strings.Join(escapedSlice, ",")
}

// QuoteSQLName 对SQL列/表等对象名称进行转义，一定会再前后加反引号
// 对于 "tbl.name" 会被处理为 "`tbl`.`name`"
// 注意：该函数与 GORM V1 的 DB.Table 函数不兼容，原因是 Table 会直接再输入的表名前后加反引号
// 强烈建议升级至 GORM V2
func QuoteSQLName(data string) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('`')
	for _, c := range data {
		switch c {
		case '`':
			buf.WriteString("``")
		case '.':
			buf.WriteString("`.`")
		default:
			buf.WriteRune(c)
		}
	}
	buf.WriteByte('`')
	return buf.String()
}

func Allow(data string, allow ...string) bool {
	return contains(data, allow...)
}

func contains(data string, allow ...string) bool {
	for _, item := range allow {
		if item == data {
			return true
		}
	}
	return false
}

func isDigit(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
