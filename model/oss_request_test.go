package model

import (
	"fmt"
	"testing"
)

func TestDownLoad(t *testing.T) {

	client := NewOssClient("http://oss-cn-beijing.aliyuncs.com", "LTAI5tEAzFMCX559VD8mRDoZ", "5IV7ZpVx95vBHZ3Y74jr9amaMtXpCQ")
	bucket, _ := client.Bucket("apipost")
	//bucket.GetConfig()
	list, _ := bucket.ListObjects()
	fmt.Println("1231231", list.Objects)
	DownLoad(client, "kunpeng/test/5dd5aa9d-12ef-4510-95fe-fd915aca8dac.csv", "D:\\123\\report.csv", "apipost")
}
